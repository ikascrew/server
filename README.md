# ikascrew server

ikascrew(VJシステム)の映像サーバ。
OpenCV のウィンドウに映像を描画し、gRPC 経由で切り替え・エフェクト操作を受け付ける。

## 使い方

事前に ikasbox からプロジェクト情報を取得してワークファイルを作成する。
実行時に ikasbox は不要(ワークファイルのみ参照)。

```
go run ./cmd/ika-server create <project-id>   # ワーク(.server/config.json)を作成 ※ikasbox が必要
go run ./cmd/ika-server start                 # サーバ起動 ※ikasbox 不要
```

### 終了方法

フルスクリーン中はプレイ中とみなし、終了操作(ESC / ☓ボタン / Ctrl+C)をすべて無視する。
終了するには gRPC の Sync でフルスクリーンを解除(トグル)してから、ESC / ☓ / Ctrl+C のいずれかを行う。

## TODO / 設計メモ

クライアントからの一覧取得をワークファイル方式に統一(クライアント側も .client を利用中)

### Next() のパニック捕捉(safeNext)

2026-07-26 の作業ツリー消失で失われ、まだ作り直していない。

- `stream.go` の `Stream.Get` は 3 本のビデオの `Next()` を goroutine から直接呼んでいる。プラグインが `Next()` でパニックすると server ごと落ちる
- 消失前は `safeNext` で `recover()` し、エラーとして返していた(`stream_test.go` の `TestStreamGetPropagatesNextError` のコメントに名前だけ残っている)
- goroutine 内のパニックは呼び出し側で捕まえられないので、各 goroutine の中で `recover()` する必要がある

### エフェクトの設計変更

現状 AddWeighted でのクロスフェード(switch)しかないため、分離して拡張する。

- effect -> 現状の Wait や Light -> 個別にビデオにかけられるようにする
- transition -> 現行の switch

Stream はサーバ固有の Video で、effect を持った Video 同士の transition に利用し、
クライアントで Video + effect を作成し、transition で切り替える。

クライアントは次のビデオ作成と push を行って、transition はマニュアルか任せる。

### transition 候補(OpenCV)

- マスク処理によるワイプ(CopyToWithMask + 矩形/円形マスク)
- BitwiseAnd(), BitwiseOr(), BitwiseNot(), BitwiseXor() + WithMask
- 輝度キー合成(Threshold -> CopyToWithMask)
- モーフカットのようなエフェクト

### effect 候補(OpenCV)

- 色相シフト(CvtColor HSV)
- ネガ反転(BitwiseNot)
- 残像フィードバック(前フレームと AddWeighted)
- 色収差(Split / Merge + ずらし)
- 波形歪み(Remap)
