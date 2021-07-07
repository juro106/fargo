---
title: 'markdown→htmlのテスト'
description: 'markdown から html への変換テスト'
tags: ["cms", "markdown"]
datePublished: '2021-06-30T15:42:56+09:00'
dateModified: '2021-07-04T04:54:48+09:00'
summary:
  コードチェック
  <pre><code>$ echo "hello"</code></pre>
  あいうえおかきくけこさしすせそたちつてと
  abcdefghijklmnopqrstuvwxyz
---

コード
<pre><code>$ echo "hello"</code></pre>

``` javascript
'use strict';
const main = () => {

  const list = ['hoge', 'fuga', 'piyo'];
  const hoge = list.forEach(v => {
    if (v == 'hoge') {
      return v;
    }
  });

  const greet = (arg) => {
    const greet = `Hello!! ${arg}`;
    return greet;
  }
  
  greet(hoge);
} 

document.addEventListener('DOMContentLoaded', main);
```
```
  words := "abcdefghijklmnopqrstuvwxyz"
  var number = 1234567890
```


## リスト 見出し2

1. 番号リスト
2. 番号リスト
3. 番号リスト

- 直接 HTML をかけるようにしておきたい 
  - テーブルはどっちでも良いがあればあったで良い
    - リストは綺麗にインデントされてほしい
       

### 太字

- _hello_ ← 斜体
- __hello__ ← 太字
- *hello* ← 斜体
- **hello** ← 太字

### テーブル

左寄せ
|country|language|
|:--|:--|
|日本 Japana|日本語 Japanese|
|アメリカ US|English|
|イギリス UK|English|
|中国 China|中国語 Chinese|
|スペイン ES|スペイン語 español|

右寄せ
|country|language|
|--:|--:|
|日本 Japana|日本語 Japanese|
|アメリカ US|English|
|イギリス UK|English|
|中国 China|中国語 Chinese|
|スペイン ES|スペイン語 español|

中央寄せ
|country|language|
|:-:|:-:|
|日本 Japana|日本語 Japanese|
|アメリカ US|English|
|イギリス UK|English|
|中国 China|中国語 Chinese|
|スペイン ES|スペイン語 español|

### 引用

> 引用
> これは
> 引用です。

#### 直接 HTML 

##### div

<div class="hey">
  areareare
</div>

  code?
  hello?

##### 引用


<blockquote class="hoge">
これも引用
</blockquote>
