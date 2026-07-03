import { Bell, AlertTriangle, TrendingUp, TrendingDown, Activity } from 'lucide-react'
import { useAlerts } from '../hooks/useAlerts'

const typeIcons: Record<string, React.ReactNode> = {
  volume_surge: <Activity size={14} />,
  price_spike: <TrendingUp size={14} />,
  limit_up: <TrendingUp size={14} />,
  limit_down: <TrendingDown size={14} />,
  macd_divergence: <AlertTriangle size={14} />,
}

const severityColors: Record<string, string> = {
  info: 'var(--accent-blue)',
  warning: '#f59e0b',
  critical: 'var(--down)',
}

const typeLabels: Record<string, string> = {
  volume_surge: '放量',
  price_spike: '急涨/急跌',
  limit_up: '涨停',
  limit_down: '跌停',
  macd_divergence: '背离',
}

export function AlertPanel() {
  const { alerts } = useAlerts()

  const recentAlerts = alerts.slice(-12).reverse()

  return (
    <div className="card alert-panel">
      <div className="card-title" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <span>
          <Bell size={12} style={{ marginRight: 4, verticalAlign: 'middle' }} />
          异常波动
        </span>
        {alerts.length > 0 && (
          <span className="alert-badge">
            {alerts.length}
          </span>
        )}
      </div>

      {recentAlerts.length === 0 ? (
        <div style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          padding: '24px 0',
          color: 'var(--text-tertiary)',
          fontSize: 13,
        }}>
          暂无告警
        </div>
      ) : (
        <div className="alert-list">
          {recentAlerts.map((alert, i) => (
            <div
              key={i}
              className="alert-item"
              style={{
                borderTop: i === 0 ? '1px solid var(--border-subtle)' : 'none',
                borderTopWidth: i === 0 ? 1 : 0,
              }}
            >
              <div className="alert-item-header">
                <span
                  className="alert-severity-dot"
                  style={{ background: severityColors[alert.severity] || 'var(--accent-blue)' }}
                />
                <span className="alert-code">{alert.code}</span>
                <span className="alert-type-tag">
                  {typeIcons[alert.type]}
                  <span style={{ marginLeft: 4 }}>{typeLabels[alert.type] || alert.type}</span>
                </span>
                <span className="alert-time">
                  {new Date(alert.time).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })}
                </span>
              </div>
              <div className="alert-message">{alert.message}</div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
