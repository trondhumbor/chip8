# CHIP8 emulator in Go

![CHIP8 screenshot](/screenshots/chip8.png?raw=true "CHIP8 screenshot")

This is a fairly basic CHIP8 emulator in Go, with the rendering handled by the [pixel/pixelgl](https://github.com/faiface/pixel/) library.

```usage: chip8 romfile```


Thanks to:
* [Cowgod's technical reference](http://devernay.free.fr/hacks/chip8/C8TECH10.HTM) for providing a good overview of the CHIP8 architecture
* [Jackson Sommerich' blogpost](https://jackson-s.me/2019/07/13/Chip-8-Instruction-Scheduling-and-Frequency.html) for accurate instruction timings

Known issues/future work:
* While most games work fine, there's some that aren't fully compatible yet
* Audio support