export function TitleBar() {
  return (
    <div className="title-bar">
      <div className="title-bar-text">A 股智能助手</div>
      <div className="title-bar-controls">
        <button className="title-bar-btn" onClick={() => window.electronAPI?.minimize()}>─</button>
        <button className="title-bar-btn" onClick={() => window.electronAPI?.toggleMaximize()}>☐</button>
        <button className="title-bar-btn title-bar-close" onClick={() => window.electronAPI?.close()}>✕</button>
      </div>
    </div>
  )
}
