
ikasboxと統合

 クライアントからの一覧取得をgRPC化
 ※ファイルがいけるか？

 headlessモードを作成（DBのみ使用する場合）
   -> リクエストがあれば作成？


opencv

AddWeighted()

BitwiseAnd(),BitwiseOr(),BitwiseNot(),BitwiseXor() + WithMask

マスク処理

BatchDistance()

BorderInterpolate()

CalcCovarMatrix()

CartToPolar()

現状AddWeighted()でのswitch処理しかないエフェクトを
モーフカットのようなエフェクトなどを行えるようにする

effect -> 現状のWaitやLight -> 個別にビデオにかけれるようにする
transition -> 現行のswitch

Stream はサーバ固有のVideoで、effectを持ったVideo同士のtransitionに利用し、
クライアントでVideo + effectを作成し、transitionで切り替える

クライアントは次のビデオ作成とpushを行って、transitionはマニュアルか任せる
マニュアルの場合
