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

const basicFragmentShaderSrc = `
precision mediump float;
uniform sampler2D uTexture;
uniform vec3 uColor;
varying vec2 vTexCoord;
void main() {
    vec4 texColor = texture2D(uTexture, vTexCoord);
    gl_FragColor = vec4(texColor.rgb * uColor, texColor.a);
}
`
