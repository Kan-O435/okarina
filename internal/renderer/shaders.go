package renderer

// basicVertexShaderSrc / basicFragmentShaderSrc は、テクスチャを使わず
// 単色でメッシュを描画するための最小限のシェーダー(WebGL1 / GLSL ES 1.00)。
const basicVertexShaderSrc = `
attribute vec3 aPosition;
uniform mat4 uMVP;
void main() {
    gl_Position = uMVP * vec4(aPosition, 1.0);
}
`

const basicFragmentShaderSrc = `
precision mediump float;
uniform vec3 uColor;
void main() {
    gl_FragColor = vec4(uColor, 1.0);
}
`
