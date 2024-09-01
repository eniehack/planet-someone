# planet-somwone

## 構造

このプログラムは3つのバイナリを組み合わせて動作します:

1. `planetctl`: このプログラムで用いるDBを管理するCLIツール
2. `picker`: 設定したサイトへアクセスし、新着投稿をDBへ格納するCLIツール
3. `hb`(`html-builder`): pickerが収集したDBにあるデータからHTMLを生成するCLIツール（静的サイトジェネレータ）

上記のプログラムのうち、`picker`と`hb`はcronやsystemd-timerdで定期実行させることを想定しています。

## install

1. git clone && cd working directory
2. `make build`

## example

[planet eniehack](https://eniehack.net/~eniehack/planet/)