---
title: '連番ファイル作成'
description: 'bash(shell)で連番ファイル作成'
tags: ["bash", "linux"]
datePublished: '2021-06-28T03:31:53+09:00'
dateModified: '2021-06-28T03:35:53+09:00'
code: true 
---

毎度調べてしまうのでメモ

```bash
$ for i in {001..999}; do cp 000.jpg $i.jpg; done
```

画像ファイルのときなどに使う。

