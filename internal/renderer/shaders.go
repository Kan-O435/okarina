package renderer

// basicVertexShaderSrc / basicFragmentShaderSrc は、テクスチャ(sampler2D)と
// 単色(uColor)を掛け合わせて描画する共通シェーダー(WebGL1 / GLSL ES 1.00)。
// テクスチャを持たないオブジェクトは1x1の白テクスチャ(Context.WhiteTexture)を
// バインドすることで、実質uColorだけで色が決まるようにしている。
const basicVertexShaderSrc = `
attribute vec3 aPosition;
attribute vec2 aTexCoord;
uniform mat4 uMVP;
varying vec2 vTexCoord;
void main() {
    gl_Position = uMVP * vec4(aPosition, 1.0);
    vTexCoord = aTexCoord;
}
`

// uAlphaは、雷の閃光・白フェード(internal/renderer/overlay.go参照)のように、
// テクスチャのアルファ値とは別に描画全体の不透明度を外部からアニメーション
// させたい場合だけに使うuniform。それ以外の通常描画では常に1.0を明示的に
// 設定する(Scene.Render/HUD.Render参照。WebGLのuniform初期値は0.0のため、
// 設定を怠ると全描画が透明になってしまう点に注意)。
const basicFragmentShaderSrc = `
precision mediump float;
uniform sampler2D uTexture;
uniform vec3 uColor;
uniform float uAlpha;
varying vec2 vTexCoord;
void main() {
    vec4 texColor = texture2D(uTexture, vTexCoord);
    gl_FragColor = vec4(texColor.rgb * uColor, texColor.a * uAlpha);
}
`
