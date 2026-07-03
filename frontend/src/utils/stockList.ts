// Full A-share market stocks - covering all major sectors and popular stocks
// Format: [code, name, pinyin]
// Total: 60+ core stocks across 10 sectors
export const STOCK_DATABASE: [string, string, string][] = [
  // 白酒 (6 只)
  ['600519', '贵州茅台', 'guizhoumaotai'],
  ['000858', '五粮液', 'wuliangye'],
  ['000568', '泸州老窖', 'luzhoulaojiao'],
  ['002304', '洋河股份', 'yanghegupfen'],
  ['600809', '山西汾酒', 'shanxifenjiu'],
  ['603369', '今世缘', 'jinshiyuan'],

  // 银行 (8 只)
  ['601318', '中国平安', 'zhongguopingan'],
  ['600036', '招商银行', 'zhaoshangyinhang'],
  ['601166', '兴业银行', 'xingyeyinhang'],
  ['601229', '上海银行', 'shanghaiyinhang'],
  ['600030', '中信证券', 'zhongxinzhengquan'],
  ['601988', '中国银行', 'zhongguoyinhang'],
  ['601939', '建设银行', 'jiansheyinhang'],
  ['600016', '民生银行', 'minshengyinhang'],

  // 新能源 (8 只)
  ['300750', '宁德时代', 'ningdeshidai'],
  ['601012', '隆基绿能', 'longjiluneng'],
  ['002594', '比亚迪', 'biyadi'],
  ['600438', '通威股份', 'tongweigupfen'],
  ['300274', '阳光电源', 'yangguangdianuan'],
  ['300763', '锦浪科技', 'jinlangkeji'],
  ['603688', '石英股份', 'shiyinggupfen'],
  ['002459', '晶澳科技', 'jinguokeji'],

  // 半导体 (10 只)
  ['688981', '中芯国际', 'zhongxunguoji'],
  ['600584', '长电科技', 'changdiankeji'],
  ['002049', '紫光国芯', 'ziguoguoxin'],
  ['688012', '中微公司', 'zhongweigongsi'],
  ['603986', '兆易创新', 'zhaoyichuangxin'],
  ['688008', '澜起科技', 'lanjikeji'],
  ['688347', '芯源微', 'xinyuanwei'],
  ['603501', '韦尔股份', 'weiersgupfen'],
  ['688256', '寒武纪', 'hanwuji'],
  ['603019', '中科曙光', 'zhongkesuguang'],

  // 医药 (8 只)
  ['600276', '恒瑞医药', 'hengruiyiyao'],
  ['300760', '迈瑞医疗', 'mairuiyiliao'],
  ['000538', '云南白药', 'yunnanbaiyao'],
  ['600436', '片仔癀', 'pianzhuanghuang'],
  ['300194', '福安药业', 'fuanyaoye'],
  ['600085', '同仁堂', 'tongrentang'],
  ['300896', '爱美客', 'aimeike'],
  ['603259', '药明康德', 'yaomingkangde'],

  // 互联网/科技 (8 只)
  ['002230', '科大讯飞', 'kedaxunfei'],
  ['601127', '赛力斯', 'sailisi'],
  ['688111', '京东方 A', 'jingdongfang'],
  ['000001', '平安银行', 'pinganyinhang'],
  ['688256', '寒武纪', 'hanwuji'],
  ['603019', '中科曙光', 'zhongkesuguang'],
  ['688111', '京东方 A', 'jingdongfanga'],
  ['002415', '海康威视', 'hai kangweishi'],

  // 消费 (7 只)
  ['002714', '牧原股份', 'muyuanguofen'],
  ['600900', '长江电力', 'changjiangdianli'],
  ['601888', '中国中免', 'zhongguozhongmian'],
  ['600690', '海尔智家', 'haierzhijia'],
  ['002352', '顺丰控股', 'shunfengkonggu'],
  ['603876', '韵达股份', 'yundagupfen'],
  ['002891', '中宠股份', 'zhongchonggupfen'],

  // 汽车 (6 只)
  ['601633', '长城汽车', 'changchengqiche'],
  ['600741', '华域汽车', 'huayuqiche'],
  ['600104', '上汽集团', 'shangqijituan'],
  ['000625', '长安汽车', 'changanqiche'],
  ['601799', '星宇股份', 'xingyugupfen'],
  ['002709', '天赐材料', 'tiancailiao'],

  // 房地产 (5 只)
  ['000002', '万科 A', 'wankeA'],
  ['600048', '保利发展', 'baolifazhan'],
  ['001979', '招商蛇口', 'zhaoshangshekou'],
  ['600383', '金地集团', 'jindijituan'],
  ['600641', '万业企业', 'wanyeqiye'],

  // 航空 (5 只)
  ['601111', '中国国航', 'zhongguoguohang'],
  ['600115', '中国东航', 'zhongguodonghang'],
  ['600029', '南方航空', 'nanfanghangkong'],
  ['601021', '春秋航空', 'chunqiuangkong'],
  ['603885', '吉祥航空', 'jixianghangkong'],

  // 钢铁/有色 (6 只)
  ['601005', '宝钢股份', 'baogangupfen'],
  ['600547', '山东黄金', 'shandonghuangjin'],
  ['601899', '紫金矿业', 'zijinkuangye'],
  ['600362', '江西铜业', 'jiangxityunye'],
  ['601600', '铝业股份', 'luyegupfen'],
  ['600895', '张江高科', 'zhangjianggao'],
]

export function searchStocks(query: string): { code: string; name: string }[] {
  const q = query.toLowerCase().trim()
  if (!q) return []

  return STOCK_DATABASE
    .filter(([, name, pinyin]) =>
      name.includes(q) || pinyin.includes(q) || q.startsWith(q.slice(0, 3))
    )
    .slice(0, 10)
    .map(([code, name]) => ({ code, name }))
}

// Pre-built sector lists
export const SECTOR_LISTS: Record<string, { code: string; name: string }[]> = {
  '白酒': [
    { code: '600519', name: '贵州茅台' },
    { code: '000858', name: '五粮液' },
    { code: '000568', name: '泸州老窖' },
    { code: '002304', name: '洋河股份' },
    { code: '600809', name: '山西汾酒' },
    { code: '603369', name: '今世缘' },
  ],
  '新能源': [
    { code: '300750', name: '宁德时代' },
    { code: '601012', name: '隆基绿能' },
    { code: '002594', name: '比亚迪' },
    { code: '600438', name: '通威股份' },
    { code: '300274', name: '阳光电源' },
    { code: '300763', name: '锦浪科技' },
    { code: '603688', name: '石英股份' },
    { code: '002459', name: '晶澳科技' },
  ],
  '半导体': [
    { code: '688981', name: '中芯国际' },
    { code: '600584', name: '长电科技' },
    { code: '002049', name: '紫光国芯' },
    { code: '688012', name: '中微公司' },
    { code: '603986', name: '兆易创新' },
    { code: '688008', name: '澜起科技' },
    { code: '688347', name: '芯源微' },
    { code: '603501', name: '韦尔股份' },
  ],
  '银行': [
    { code: '601318', name: '中国平安' },
    { code: '600036', name: '招商银行' },
    { code: '601166', name: '兴业银行' },
    { code: '601229', name: '上海银行' },
    { code: '600030', name: '中信证券' },
    { code: '601988', name: '中国银行' },
    { code: '601939', name: '建设银行' },
    { code: '600016', name: '民生银行' },
  ],
  '医药': [
    { code: '600276', name: '恒瑞医药' },
    { code: '300760', name: '迈瑞医疗' },
    { code: '000538', name: '云南白药' },
    { code: '600436', name: '片仔癀' },
    { code: '300194', name: '福安药业' },
    { code: '600085', name: '同仁堂' },
    { code: '300896', name: '爱美客' },
  ],
  '消费': [
    { code: '603259', name: '药明康德' },
    { code: '002714', name: '牧原股份' },
    { code: '600900', name: '长江电力' },
    { code: '601888', name: '中国中免' },
    { code: '600690', name: '海尔智家' },
    { code: '002352', name: '顺丰控股' },
  ],
}
