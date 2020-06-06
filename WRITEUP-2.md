# HackDalton: CoolCPU 2
> Warning! There are spoilers ahead

> You'll want to get through [the writeup of CoolCPU 1](./WRITEUP-1.md) first.

The main new thing in part 2 is the addition of the external hard drive, or, as the website calls it, the DynamicBlast! Engine. We're told that one of the sectors has our flag, so we need to write some code that scans each sector for a flag.

Also, as described on the website, the Engine has some strict timing requirements.

> **Sidenote:**
> These requirements are supposed to simulate certain types of peripherals. This was a common trick back when CPU processing power was more expensive: make your hardware dumb and make the CPU do the work.
>
> Specifically, one famous example of this was the floppy disk drive of the [Apple \]\[](https://en.wikipedia.org/wiki/Apple_II) computer. Rather than build a floppy disk controller into the drive, they made the CPU do all the work, which required very precise timing, since the CPU's instructions had to align with the movements of the floppy disk.
>
> (this is a bit of a simplification&mdash;if you're interested, you can watch [this video about the Apple\]\['s drive](https://www.youtube.com/watch?v=w3VZFhNQRmU))

So, let's break down the steps we need to do:
1. Tell the Engine to start copying a sector.
2. Do the pokes at the correct time, as described on the website.
3. Check if the sector had anything.
4. If it did, print it out! Otherwise, go back to step 1, copying the next sector.

How do we know how to start copying a sector? The website tells us the exact things we need to do. So, let's start our code like that:
```
	CON 0x90 ; we will copy to 0x90
	STA 0xF2 ; write the destination address
	CON 0 ; start register A at sector 0
start_copy_sector:
	STA 0xF3 ; write the source sector

	; now we need to start the copy
	; we want to save register A though
	; so we swap it, and then start the copy
	; at the end we'll swap it back
	SWB
	CON 1
	STA 0xF4 ; start the copy
```

(like in the [first writeup](./WRITEUP-1.md), we'll write the code using the instruction names, and then convert it to hex at the end. We're also using the `;` character to designate comments)

Now we have to do the pokes.

```
	CON 5 ; we'll now use register A to track the poke index
poke_loop:
	STA 0xF5			; 1 cycle
	DEC				; 2 cycles
	JNZ poke_loop			; 3 cycles
```

But wait! The website told us that we need 36 cycles between the pokes (in this case, that's the `STA 0xF5` instruction). So, we need to add some code to slow down the CPU. Let's do that now:

```
	CON 5 ; we'll now use register A to track the poke index
poke_loop:
	STA 0xF5			; 1 cycle

	; this part is just to delay things
	SWC				; 1 cycle
	CON 5				; 1 cycle
delay_loop:
	DEC				; 2 cycles
	JNZ	delay_loop		; 3 cycles
	SWC				; 1 cycle
	NOP				; 1 cycle
	NOP				; 1 cycle

	DEC				; 2 cycles
	JNZ poke_loop			; 3 cycles
```

That might be a little tricky to understand, but that new chunk of code will slow things down enough so that you get the right timing. (you're encouraged to walk through the instructions yourself, counting the cycles as you go)

> **Alternative**: If the whole delay_loop thing confuses you, keep in mind you really just need to slow down the program. Another valid way of doing this:
> ```
>	CON 5 ; we'll now use register A to track the poke index
>poke_loop:
>	STA 0xF5		; 1 cycle
>
>	NOP				; 1 cycle
>	NOP				; 1 cycle
>	NOP				; 1 cycle
>	NOP				; 1 cycle
>	NOP				; 1 cycle
>	NOP				; 1 cycle
>	NOP				; 1 cycle
>	NOP				; 1 cycle
>	NOP				; 1 cycle
>	NOP				; 1 cycle
>	NOP				; 1 cycle
>	NOP				; 1 cycle
>	NOP				; 1 cycle
>	NOP				; 1 cycle
>	NOP				; 1 cycle
>	NOP				; 1 cycle
>	NOP				; 1 cycle
>	NOP				; 1 cycle
>	NOP				; 1 cycle
>	NOP				; 1 cycle
>	NOP				; 1 cycle
>	NOP				; 1 cycle
>	NOP				; 1 cycle
>	NOP				; 1 cycle
>	NOP				; 1 cycle
>	NOP				; 1 cycle
>	NOP				; 1 cycle
>	NOP				; 1 cycle
>	NOP				; 1 cycle
>	NOP				; 1 cycle
>
>	DEC				; 2 cycles
>	JNZ poke_loop	; 3 cycles
> ```
> (that's a total of 30 NOPs! so, 36 cycles)

Ok, so now we have our pokes. Going back to the list of steps, we now need to check if the sector had anything in it. Since we copied it to 0x90, we can just load whatever's at 0x90, and check if it's 0.

```
	LDA 0x90
	JNZ print ; if it's not 0, print it out

	; if we're still here, then we need to go to the next sector
	; we swap A with B in order to get back our sector count
	SWB
	INC ; next sector
	JP start_copy_sector
```

For the `print` code, we can reuse what we did in [CoolCPU 1](./WRITEUP-1.md), as long as you remember to change the start address of the print from `0x93` to `0x90`:

```
print:
	CON 0x90
print_loop:
	LDB [A]
	SWB
	JZ end
	SWB
	STB 0xF1
	INC
	JP print_loop
end:
	HCF
```

## Stitching it together
Finally, we can put all the pieces together to get the result:
```
	CON 0x90 ; we will copy to 0x90
	STA 0xF2 ; write the destination address
	CON 0 ; start register A at sector 0
start_copy_sector:
	STA 0xF3 ; write the source sector

	; now we need to start the copy
	; we want to save register A though
	; so we swap it, and then start the copy
	; at the end we'll swap it back
	SWB
	CON 1
	STA 0xF4 ; start the copy


	CON 5 ; we'll now use register A to track the poke index
poke_loop:
	STA 0xF5			; 1 cycle

	; this part is just to delay things
	SWC				; 1 cycle
	CON 5				; 1 cycle
delay_loop:
	DEC				; 2 cycles
	JNZ	delay_loop		; 3 cycles
	SWC				; 1 cycle
	NOP				; 1 cycle
	NOP				; 1 cycle

	DEC				; 2 cycles
	JNZ poke_loop			; 3 cycles

	LDA 0x90
	JNZ print ; if it's not 0, print it out

	; if we're still here, then we need to go to the next sector
	; we swap A with B in order to get back our sector count
	SWB
	INC ; next sector
	JP start_copy_sector

print:
	CON 0x90
print_loop:
	LDB [A]
	SWB
	JZ end
	SWB
	STB 0xF1
	INC
	JP print_loop
end:
	HCF
```

And that's our assembly code! Now, you need to convert it to hex. You could do this manually, but that can get quite tricky, especially with keeping track of the different labels. So, you might find it useful to use a tool like [customasm](https://hlorenzi.github.io/customasm/web/), or modify [an existing, similar assembler](https://github.com/thatoddmailbox/gbasm) to work with the CoolCPU.

For reference, the assembled version might look like this:
```
 outp | addr | data

  0:0 |    0 | 22 90          ; CON 0x90
  2:0 |    2 | 11 f2          ; STA 0xF2
  4:0 |    4 | 22 00          ; CON 0
  6:0 |    6 |                ; start_copy_sector:
  6:0 |    6 | 11 f3          ; STA 0xF3
  8:0 |    8 | 40             ; SWB
  9:0 |    9 | 22 01          ; CON 1
  b:0 |    b | 11 f4          ; STA 0xF4
  d:0 |    d | 22 05          ; CON 5
  f:0 |    f |                ; poke_loop:
  f:0 |    f | 11 f5          ; STA 0xF5
 11:0 |   11 | 41             ; SWC
 12:0 |   12 | 22 05          ; CON 5
 14:0 |   14 |                ; delay_loop:
 14:0 |   14 | 21             ; DEC
 15:0 |   15 | 32 14          ; JNZ	delay_loop
 17:0 |   17 | 41             ; SWC
 18:0 |   18 | 00             ; NOP
 19:0 |   19 | 00             ; NOP
 1a:0 |   1a | 21             ; DEC
 1b:0 |   1b | 32 0f          ; JNZ poke_loop
 1d:0 |   1d | 01 90          ; LDA 0x90
 1f:0 |   1f | 32 25          ; JNZ print
 21:0 |   21 | 40             ; SWB
 22:0 |   22 | 20             ; INC
 23:0 |   23 | 30 06          ; JP start_copy_sector
 25:0 |   25 |                ; print:
 25:0 |   25 | 22 90          ; CON 0x90
 27:0 |   27 |                ; print_loop:
 27:0 |   27 | 05             ; LDB [A]
 28:0 |   28 | 40             ; SWB
 29:0 |   29 | 31 31          ; JZ end
 2b:0 |   2b | 40             ; SWB
 2c:0 |   2c | 12 f1          ; STB 0xF1
 2e:0 |   2e | 20             ; INC
 2f:0 |   2f | 30 27          ; JP print_loop
 31:0 |   31 | ff             ; HCF
```

And so, in the simulator, you could write: 
```
22 90
11 f2
22 00

11 f3
40
22 01
11 f4
22 05

11 f5
41
22 05

21
32 14
41
00
00
21
32 0f
01 90
32 25
40
20
30 06

22 90

05
40
31 31
40
12 f1
20
30 27
ff
```