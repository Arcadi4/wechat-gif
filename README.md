# wechat-gif

中文 | [EN](README_EN.md)

一个使得 gif 可以在微信中以图片（表情包）形式发送的简单压缩工具。

压缩指定的gif:

```bash
wechat-gif path/to/gif1 path/to/gif2 [...]
```

压缩指定目录下所有的gif:

```bash
wechat-gif -d path/of/gifs
```

压缩至可以自动播放的大小:

```bash
wechat-gif -a path/to/gif1
```

## 原理

使用 macOS 平台微信 3.8.10 版本进行测试，以图片形式发送 gif 的要求为：

- 宽高均不超过 1000px
- 文件大小不超过 5MiB

并且，如果要在其他客户端自动播放：

- 文件大小不超过 1MiB

程序通过降低分辨率来将输入图像压缩至微信可以接受的尺寸和文件大小。
