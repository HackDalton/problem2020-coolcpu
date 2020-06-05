package cpu

import (
	"context"
	"errors"
	"fmt"
)

const DefaultBankSize = 0x80

var errOnFire = errors.New("cpu: on fire")

type CPU struct {
	ROM [DefaultBankSize]uint8
	RAM [DefaultBankSize]uint8

	WriteCallback func(c uint8)

	PC uint8

	A uint8
	B uint8
	C uint8
}

func NewCPU() *CPU {
	c := &CPU{}
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

	panic(fmt.Errorf("cpu: tried to write to out of bounds address 0x%x", address))
}

func (c *CPU) Step() {
	instruction := c.Read(c.PC)
	instructionSize := uint8(1)
	skipPCIncrement := false

	switch instruction {
	case 0x00:
		// NOP, do nothing

	case 0x01:
		// LDA <data>
		c.A = c.Read(c.PC + 1)
		instructionSize = 2

	case 0x02:
		// LDB <data>
		c.B = c.Read(c.PC + 1)
		instructionSize = 2

	case 0x03:
		// LDC <data>
		c.C = c.Read(c.PC + 1)
		instructionSize = 2

	case 0x04:
		// LDA [A]
		c.A = c.Read(c.A)

	case 0x05:
		// LDB [A]
		c.B = c.Read(c.A)

	case 0x06:
		// LDC [A]
		c.C = c.Read(c.A)

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

	case 0x15:
		// STB [A]
		c.Write(c.A, c.B)

	case 0x16:
		// STC [A]
		c.Write(c.A, c.C)

	case 0x20:
		// INC
		c.A += 1

	case 0x21:
		// DEC
		c.A -= 1

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

	case 0x31:
		// JZ <data>
		destination := c.Read(c.PC + 1)
		instructionSize = 2
		if c.A == 0 {
			c.PC = destination
			skipPCIncrement = true
		}

	case 0x32:
		// JNZ <data>
		destination := c.Read(c.PC + 1)
		instructionSize = 2
		if c.A != 0 {
			c.PC = destination
			skipPCIncrement = true
		}

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
