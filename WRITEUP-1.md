# HackDalton: CoolCPU 1
> Warning! There are spoilers ahead

The two sample programs on the CoolCPU website help you understand the basics of how writing code works, so you should go through those two first.

So, from the description, we know that there's a string of characters at `0x93` that we want to print out. And, we know it's null-terminated (that is, it ends in a `0x00` bytes).

The basic idea for writing code to print the flag is to, starting at `0x93`, read each byte and print it out, until we hit that `0x00`. The instructions to do that might look something like this:
```
	CON 0x93				; 22 93
loop: 			; this will be at position 0x02
	LDB [A]					; 05
	SWB						; 40
	JZ end					; 31 0C
	SWB						; 40
	STB 0xF1				; 12 F1
	INC						; 20
	JP loop					; 30 02
end:			; this will be at position 0x0C
	HCF						; FF
```
The comments on the right show you the hexadecimal representations of each instruction.

Let's go through this step by step: first, the `CON 0x93` sets A to `0x93`, which is where we start our loop. We're using A to keep track of the current memory address.

Then, we start our actual loop. We use `LDB [A]` to read the value at A (the next character of our flag) and store it in register B.

Next up, we need to check if the byte we read was the `0x00` that signals the end. Unfortunately, we're only able to make comparisons with register A. So, we do `SWB` to swap registers A and B, `JZ` to check if it's zero, and then another `SWB` to bring things back to normal.

If it _was_ zero, the `JZ` would jump to the `HCF` instruction that ends the program, and we're done. If it _wasn't_, then the `STB 0xF1` writes the character to the output. Then the `INC` adds 1 to register A, and the `JP loop` goes back to the start of the loop so we can print the next character.

Combining the hexadecimal representations of each instruction, we get `22 93 05 40 31 0C 40 12 F1 20 30 02 FF`. Running this code in the simulator, we get our flag!