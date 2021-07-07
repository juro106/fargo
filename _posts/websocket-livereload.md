---
datePublished: '2021-07-06T22:22:37+09:00'
dateModified: '2021-07-07T22:35:39+09:00'
title: 'WebSocket と Live Reload'
description: 'Live Reload を実現するためには？'
tags:
  - cms
  - go
  - javascript
summary: |-

---

色々とやり方はあるようだけれども、自分の環境に合うやつというのは案外なかったりする。

- 静的サイトジェネレーターの下書きのときに使いたい
  - 記事を書いているときジェネレーターを作っているときではない。
  - markdown の1ファイルを監視するのとは多分ちょっと違う
- ローカルサーバーが見ているのは golang の Map
  - deploy 先のディレクトリ（`_site`とか`dist`）以下のファイル群ではない
  - 各ページの HTML にリロード用の javascript を埋め込むため

結局 HUGO のような仕組みを作るしかなさそう…

1. ファイルの変更を監視 ← local server
2. 変更があったら build する ← local server
3. build されたら reload する ← client

簡単に出来そうで出来ない。

ファイルに変更から build への流れが意外と難しい。build が2回呼ばれたりしてエラーになる。

## 仕組み（机上の空論）

- 1度 build() が呼ばれたら、他に何かイベントが発生してもスルー。
- build() が呼ばれたら何かしらのフラグを立てる building など
- build() が呼ばれたら何かしらのフラグを立てる building など

これで良いのかわからんが動いている。




