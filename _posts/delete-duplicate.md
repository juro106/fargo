---
datePublished: '2021-07-03T21:00:53+09:00'
dateModified: '2021-07-03T23:27:47+09:00'
title: 'slice 重複削除 Golang'
description: 'Golang で2つ以上のリストの重複を調べるときなど'
tags:
  - go
summary: |-

---

2つ以上の slice を比較して重複を削除するときなど

``` go
func unique(ss ...[]string) []string {
    m := map[string]int{}
    for _, s := range ss {
        for _, v := range s {
            m[v]++ // 出現回数カウント
        }
    }
    res := []string{}
    for k, v := range m {
        if v == 1 {
            res = append(res, k) // 重複していないファイルを抽出
        }
    }
    return res
}
```

