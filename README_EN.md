# wechat-gif

[中文](README.md) | EN

A simple compression tool that enables GIFs to be sent as images (stickers) rather than files on WeChat.

Compress specified GIFs:

```bash
wechat-gif path/to/gif1 path/to/gif2 [...]
```

Compress all GIFs in a specified directory:

```bash
wechat-gif -d path/of/gifs
```

Compress to a size that allows autoplay:

```bash
wechat-gif -a path/to/gif1
```

## How Does it Work

Tested on macOS WeChat 3.8.10, the requirements for sending gif as image are:

- Width and height are both less than 1000px
- File size is less than 5MiB

Additionally, if you want the gif to be played automatically on other clients:

- File size is less than 1MiB

The program reduces the resolution to compress the input GIF to dimensions and file sizes acceptable by WeChat.
