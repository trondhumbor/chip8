# CHIP8 emulator in Go

This is a fairly basic CHIP8 emulator in Go, with the rendering handled by the pixel/pixelgl library.

```usage: chip8 romfile```


Thanks to:
* [Cowgod's technical reference](http://devernay.free.fr/hacks/chip8/C8TECH10.HTM) for providing good overview of the CHIP8 architecture
* [Jackson Sommerich' blogpost](https://jackson-s.me/2019/07/13/Chip-8-Instruction-Scheduling-and-Frequency.html) for accurate instruction timings

Known issues/future work:
* While most games work fine, there's some that needs to be looked into
* Audio support
* The keypad occasionally won't play ball on Linux, possibly a pixel/opengl issue.