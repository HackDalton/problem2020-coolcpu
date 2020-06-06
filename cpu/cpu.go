package cpu

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
)

const DefaultBankSize = 0x80
const SectorSize = 32

const dbeCycleDistance = 36

var errOnFire = errors.New("cpu: on fire")

type CPU struct {
	ROM [DefaultBankSize]uint8
	RAM [DefaultBankSize]uint8

	WriteCallback func(c uint8)
	Version       Version

	PC uint8

	A uint8
	B uint8
	C uint8

	Cycle int

	magicSector uint8
	dbeSrc      uint8
	dbeDst      uint8
	dbeStarted  bool
	dbePke      uint8
	dbePkeCycle int
}

func NewCPU(version Version) *CPU {
	c := &CPU{
		Version: version,

		magicSector: uint8(rand.Int31n(256)),
	}
	/*
		for i := 0; i < DefaultBankSize; i++ {
			c.ROM[i] = uint8(rand.Int31n(256))
			c.RAM[i] = uint8(rand.Int31n(256))
		}
	*/
	return c
}

func (c *CPU) Read(address uint8) uint8 {
	if address >= 0 && address < 0x80 {
		// read from ROM
		return c.ROM[address]
	}

	if address >= 0x80 && address < 0xF0 {
		// read from RAM
		return c.RAM[address-0x80]
	}

	panic(fmt.Errorf("cpu: tried to read from out of bounds address 0x%x", address))
}

func (c *CPU) Write(address uint8, data uint8) {
	if address >= 0 && address < 0x80 {
		// can't write to ROM!
		panic(errors.New("cpu: tried to write to ROM"))
	}

	if address >= 0x80 && address < 0xF0 {
		// write to RAM
		c.RAM[address-0x80] = data
		return
	}

	if address == 0xF1 {
		// OUTREG
		if c.WriteCallback != nil {
			c.WriteCallback(data)
		}
		return
	}

	if c.Version == Version2 {
		if address == 0xF2 {
			// DBEDST
			if c.dbeStarted {
				panic(fmt.Errorf("cpu: cannot set DBEDST while copying"))
			}

			c.dbeDst = data

			return
		}

		if address == 0xF3 {
			// DBESRC
			if c.dbeStarted {
				panic(fmt.Errorf("cpu: cannot set DBESRC while copying"))
			}

			c.dbeSrc = data

			return
		}

		if address == 0xF4 {
			// DBEGO
			if c.dbeStarted {
				panic(fmt.Errorf("cpu: cannot set DBEGO while copying"))
			}

			c.dbeStarted = true
			c.dbePke = 5
			c.dbePkeCycle = c.Cycle

			return
		}

		if address == 0xF5 {
			// DBEPKE

			if !c.dbeStarted {
				panic(fmt.Errorf("cpu: tried to poke without copy in progress"))
			}

			if c.dbePke != data {
				panic(fmt.Errorf("cpu: incorrect poke index; expected %d, got %d", c.dbePke, data))
			}

			if c.dbePke != 5 {
				// check cycles
				if c.dbePkeCycle+dbeCycleDistance < c.Cycle {
					panic(fmt.Errorf("cpu: poke of DynamicBlast Engine was too late"))
				}
				if c.dbePkeCycle+dbeCycleDistance > c.Cycle {
					panic(fmt.Errorf("cpu: poke of DynamicBlast Engine was too early"))
				}
			}

			c.dbePke--
			c.dbePkeCycle = c.Cycle

			if c.dbePke == 0 {
				// do the copy
				if c.dbeSrc == c.magicSector {
					flag := "hackDalton{p0k3_p0k3_MnejqOaw3e}"
					for i := uint8(0); i < SectorSize; i++ {
						c.Write(c.dbeDst+i, flag[i])
					}
				} else {
					for i := uint8(0); i < SectorSize; i++ {
						c.Write(c.dbeDst+i, 0)
					}
				}
				c.dbeStarted = false
			}

			return
		}
	}

	panic(fmt.Errorf("cpu: tried to write to out of bounds address 0x%x", address))
}

func (c *CPU) Step() {
	instruction := c.Read(c.PC)
	instructionSize := uint8(1)
	skipPCIncrement := false
	cycles := 1

	switch instruction {
	case 0x00:
		// NOP, do nothing

	case 0x01:
		// LDA <data>
		c.A = c.Read(c.Read(c.PC + 1))
		instructionSize = 2

	case 0x02:
		// LDB <data>
		c.B = c.Read(c.Read(c.PC + 1))
		instructionSize = 2

	case 0x03:
		// LDC <data>
		c.C = c.Read(c.Read(c.PC + 1))
		instructionSize = 2

	case 0x04:
		// LDA [A]
		c.A = c.Read(c.A)
		cycles = 2

	case 0x05:
		// LDB [A]
		c.B = c.Read(c.A)
		cycles = 2

	case 0x06:
		// LDC [A]
		c.C = c.Read(c.A)
		cycles = 2

	case 0x11:
		// STA <data>
		c.Write(c.Read(c.PC+1), c.A)
		instructionSize = 2

	case 0x12:
		// STB <data>
		c.Write(c.Read(c.PC+1), c.B)
		instructionSize = 2

	case 0x13:
		// STC <data>
		c.Write(c.Read(c.PC+1), c.C)
		instructionSize = 2

	case 0x14:
		// STA [A]
		c.Write(c.A, c.A)
		cycles = 2

	case 0x15:
		// STB [A]
		c.Write(c.A, c.B)
		cycles = 2

	case 0x16:
		// STC [A]
		c.Write(c.A, c.C)
		cycles = 2

	case 0x20:
		// INC
		c.A += 1
		cycles = 2

	case 0x21:
		// DEC
		c.A -= 1
		cycles = 2

	case 0x22:
		// CON <data>
		c.A = c.Read(c.PC + 1)
		instructionSize = 2

	case 0x30:
		// JP <data>
		destination := c.Read(c.PC + 1)
		instructionSize = 2
		c.PC = destination
		skipPCIncrement = true
		cycles = 2

	case 0x31:
		// JZ <data>
		destination := c.Read(c.PC + 1)
		instructionSize = 2
		if c.A == 0 {
			c.PC = destination
			skipPCIncrement = true
		}
		cycles = 3

	case 0x32:
		// JNZ <data>
		destination := c.Read(c.PC + 1)
		instructionSize = 2
		if c.A != 0 {
			c.PC = destination
			skipPCIncrement = true
		}
		cycles = 3

	case 0x40:
		// SWB
		temp := c.A
		c.A = c.B
		c.B = temp

	case 0x41:
		// SWC
		temp := c.A
		c.A = c.C
		c.C = temp

	case 0xFF:
		// HCF
		panic(errOnFire)

	default:
		// unknown
		panic(fmt.Errorf("cpu: tried to run illegal instruction 0x%x", instruction))
	}

	if !skipPCIncrement {
		c.PC += instructionSize
	}

	c.Cycle += cycles
}

func (c *CPU) StepSafe() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	c.Step()

	return nil
}

func (c *CPU) Run(ctx context.Context) error {
	c.PC = 0
	for {
		err := c.StepSafe()

		if err == errOnFire {
			break
		}

		if err != nil {
			return err
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
	return nil
}
