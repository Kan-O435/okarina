# 馬アセット候補(演奏で馬を呼ぶ機能用)

「演奏を流したら馬を呼んで馬に乗る」機能のために調べた、Linkの見た目
(KayKit低ポリスタイル)に合いそうな馬の3Dモデル。実装(コードへの組み込み)は
まだ行っておらず、このフォルダには参考用にアセットとプレビュー画像だけを
まとめてある。

## 採用候補: Quaternius "Ultimate Animated Animal Pack" のHorse

- 配布元: https://quaternius.com/packs/ultimateanimatedanimals.html
- ライセンス: **CC0**(パブリックドメイン、商用可・クレジット表記不要)
- 元データ: Google Driveで配布されている`glTF/Horse.gltf`
  (バイナリがbase64埋め込みの単一ファイル形式)を、このプロジェクトの
  GLBローダーが読める形式に変換したものが `horse.glb`
- テクスチャ無し(画像0枚)、8個のマテリアルの単色(baseColorFactor)のみで
  構成された低ポリスタイル。KayKit(Linkのモデル)と同じく
  「フラットシェーディングの低ポリ」系統で、違和感なく組み合わせられる見た目

### 収録アニメーション(`horse.glb`内、13種類)

```
Attack_Headbutt, Attack_Kick, Death, Eating,
Gallop, Gallop_Jump,
Idle, Idle_2, Idle_Headlow, Idle_HitReact1, Idle_HitReact2,
Jump_toIdle, Walk
```

### 依頼された3つのモーションとの対応案

| やりたいこと | 使えそうなアニメーション |
|---|---|
| 馬が来るところ(登場) | `Gallop`(疾走) |
| 馬に乗っているところ | `Idle` / `Walk`(そのまま乗馬中の待機・歩行として使う想定。専用の「騎乗」アニメは無いため、馬の上にLink(KayKit Knight)を手動で乗せるポーズを組む必要がある) |
| 馬とジャンプしているところ | `Gallop_Jump`(疾走からのジャンプ)または`Jump_toIdle` |

### プレビュー画像

poly.pizza(https://poly.pizza/m/F8HAAcLeBL 、同じくQuaternius・CC0)の
3Dビューアで、見た目確認用に撮ったスクリーンショット。

- `preview-run.jpg` — 走っているところ(Gallotに相当する見た目)
- `preview-idle.jpg` — 静止ポーズ
- `preview-jump.jpg` — ジャンプのポーズ

## 実装時の注意点(未着手の課題)

- 現状のエンジン(`internal/gltf`, `internal/renderer`)は**スキニング
  (ボーンによる頂点変形)・アニメーション再生ともに未実装**。Link
  (KayKit Knight)のGLBも同様の理由で今は静止ポーズ表示のみ。馬を実際に
  「歩く・走る・ジャンプする」ように動かすには、このアニメーション再生機能を
  先に作る必要がある(CLAUDE.md 15章のAnimation Systemに相当)
- `horse.glb`は複数メッシュ(体パーツ)+8マテリアルの構成。KayKit Knightと
  同様、`LoadSkinnedGLBMesh`(スキン付きノードのメッシュを結合する既存関数)
  で静止表示は可能なはず(未検証)
- 「馬に乗っているLink」の見た目は、専用アニメーションが無いため、Linkを
  馬の背中の位置に手動でTransform配置する(お絵かき的な「乗せる」処理)か、
  あるいはLinkを非表示にして馬だけ動かす簡易版から始めるのが良さそう

## ライセンス表記

`docs/licenses/quaternius-ultimate-animated-animals-CC0.txt` 参照。
