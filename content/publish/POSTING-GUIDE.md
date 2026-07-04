# Posting Guide — A-Share Assistant

## Status Summary

| Platform | Auto-posted | Manual Required |
|----------|------------|-----------------|
| V2EX     | No         | Yes             |
| HackerNews | No       | Yes             |

Both platforms require manual posting. The content is prepared in this directory.

---

## 1. V2EX (https://v2ex.com/)

### Option A: Web Posting (Recommended)
1. Go to https://v2ex.com/new/
2. Fill in:
   - **Title**: `仿制 A 股实时看盘系统 — Go + Electron, AI Agent 辅助开发`
   - **Body**: Copy from `v2ex-post.md` (below the "正文:" section)
   - **Tags**: `Go`, `Electron`, `React`, `独立开发`
3. Submit

### Option B: API Posting (Requires Token)
```bash
# First, get your token from https://v2ex.com/settings/applications
export V2EX_TOKEN="your_token_here"
export V2EX_MEMBER_ID="your_member_id"

curl -X POST "https://www.v2ex.com/api/v1/topics" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $V2EX_TOKEN" \
  -d '{
    "title": "仿制 A 股实时看盘系统 — Go + Electron, AI Agent 辅助开发",
    "raw_content": "<p>大家好，分享一个我最近用 AI Agent 团队辅助完成的开源项目...</p>",
    "tags": ["Go", "Electron", "React", "独立开发"]
  }'
```

**Note:** V2EX API requires a token from https://v2ex.com/settings/applications. Without it, the web form is the only option.

---

## 2. HackerNews (https://news.ycombinator.com/)

### Steps:
1. Go to https://news.ycombinator.com/submitlink
2. Fill in:
   - **Title**: `A-share Trading Assistant — Go + Electron, built solo with AI Agents`
   - **URL**: `https://github.com/chenqiangxu219-design/a-share-assistant`
3. Click Submit (HN link submissions don't have a body field)

### Tips:
- Best posting time for HN (US East): 8-10 AM or 4-6 PM
- The "Show HN" prefix is optional but helps visibility for new projects
- Consider: `Show HN: A-share Trading Assistant — Go + Electron, built solo with AI Agents`

---

## 3. Cross-posting Checklist

- [ ] V2EX post submitted
- [ ] HackerNews post submitted
- [ ] Monitor both for comments and respond promptly
- [ ] Update README with traffic sources if needed

## Key Links to Share in Comments

- **Repo**: https://github.com/chenqiangxu219-design/a-share-assistant
- **Release**: https://github.com/chenqiangxu219-design/a-share-assistant/releases/tag/v1.0.0
- **Downloads**: macOS (.dmg), Windows (.exe), Linux (.AppImage)
