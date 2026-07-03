import { HashRouter, Routes, Route, NavLink } from 'react-router-dom'
import { useEffect } from 'react'
import { LayoutDashboard, Filter, LineChart, Wallet, TrendingUp, Layers, Flame, Newspaper, Bell } from 'lucide-react'
import { useStore } from './store/store'
import { useWebSocket } from './hooks/useWebSocket'
import { useAlerts } from './hooks/useAlerts'
import { Dashboard } from './pages/Dashboard'
import { StockDetail } from './pages/StockDetail'
import { Screener } from './pages/Screener'
import { Backtest } from './pages/Backtest'
import { Sectors } from './pages/Sectors'
import { Heatmap } from './pages/Heatmap'
import { News } from './pages/News'
import { Portfolio } from './components/Portfolio'
import { ChatPanel } from './components/ChatPanel'
import { AlertPanel } from './components/AlertPanel'

const navItems = [
  { path: '/', label: '看盘', icon: LayoutDashboard },
  { path: '/sectors', label: '板块', icon: Layers },
  { path: '/heatmap', label: '热点', icon: Flame },
  { path: '/screener', label: '选股', icon: Filter },
  { path: '/backtest', label: '回测', icon: LineChart },
  { path: '/portfolio', label: '模拟盘', icon: Wallet },
  { path: '/news', label: '资讯', icon: Newspaper },
]

function App() {
  const setSelectedStock = useStore((s) => s.setSelectedStock)
  const setWatchlist = useStore((s) => s.setWatchlist)
  useWebSocket()
  const { alerts } = useAlerts()

  // Load all stocks from backend on mount
  useEffect(() => {
    fetch('/api/stocks')
      .then((res) => res.json())
      .then((data) => {
        if (data.stocks && Array.isArray(data.stocks)) {
          const codes = data.stocks.map((s: { code: string }) => s.code)
          setWatchlist(codes)
        }
      })
      .catch((err) => {
        console.warn('Failed to load stocks:', err)
      })
  }, [setWatchlist])

  return (
    <HashRouter>
      <div className="app">
        <div className="sidebar">
          <div className="sidebar-header">
            <TrendingUp size={20} />
            <span>A 股智能助手</span>
            {alerts.length > 0 && (
              <span className="alert-badge">
                {alerts.length}
              </span>
            )}
          </div>
          <nav className="sidebar-nav">
            {navItems.map((item) => {
              const Icon = item.icon
              return (
                <li key={item.path}>
                  <NavLink
                    to={item.path}
                    className={({ isActive }) => isActive ? 'active' : ''}
                    onClick={() => setSelectedStock(null)}
                    end={item.path === '/'}
                  >
                    <Icon size={16} />
                    <span>{item.label}</span>
                  </NavLink>
                </li>
              )
            })}
          </nav>
          <div className="sidebar-footer">
            <Bell size={14} />
            <span>异常波动</span>
          </div>
        </div>
        <main className="main-content">
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/sectors" element={<Sectors />} />
            <Route path="/heatmap" element={<Heatmap />} />
            <Route path="/screener" element={<Screener />} />
            <Route path="/backtest" element={<Backtest />} />
            <Route path="/portfolio" element={<Portfolio />} />
            <Route path="/news" element={<News />} />
            <Route path="/stock/:code" element={<StockDetail />} />
          </Routes>
          <AlertPanel />
        </main>
        <ChatPanel />
      </div>
    </HashRouter>
  )
}

export default App
