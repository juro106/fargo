---
datePublished: '2021-07-03T04:55:22+09:00'
dateModified: '2021-07-04T03:08:33+09:00'
title: 'cms 作成記録'
description: 'cms を作成する上で気づいたこと、プログラミングのメモ。'
tags:
  - 雑記
  - cms
  - go
summary: |-
---

go で cms を自作。自作とは言っても、結局は先人（巨人）が築き上げてきたライブラリ、モジュール、パーツの寄せ集めであり、そこまでいばれるものではない。

- [scrapbox](https://scrapbox.io) とブログが 3:7 で混ざったような静的サイトジェネレーターを作りたい。
- 表示はもちろん、 build を超高速にしたいので golang を勉強中。
- [jedie](https://github.com/mattn/jedie)を参考に、少しずつ解読していじくり回している感じ。
- リンクを張り巡らせたい。
- markdown の parser は [goldmark](https://github.com/yuin/goldmark)を使わせてもらっている。
- 普段は [hugo](https://github.com/gohugoio/hugo) も使っている。

2021-07-03
- とりあえず build 速度はかなりマシになった
- 一覧の冒頭部分を短くした。やはり単純にHTMLの量に左右される部分が大きい。
  - javascript にせよ、go(cmsのメイン) にせよ、細かい計算のコストは大したことない。
  - go の並行処理がちょっとわかってきた。ディレクトリ内のファイルを走査してデータをかき集めて Map や Slice を作るときや、最後のファイル書き込みに使うと良いっぽい。
  - 同時書き込みを防ぐための記述も少しわかった。
- scrapbox の機能にある、行をまるごとを上下に移動させるやつが気持ち良い。
  - 早速 Vim で `<C-Up>` `<C-Down>`  割り当てた。 `<C-j>` `<C-k>` はちょっともったいないので使えない。
  - 実際に文章を書くとなるとやはり Vim は色々都合が良い。
  - 気持ち良いだけで、そこまで頻繁に使うものでもないかもしれない。結局 dd でやってしまう。 
    
※

- 全文検索は自分のために欲しい

