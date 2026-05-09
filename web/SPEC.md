# OpenCode WebChat Frontend Specification

## 1. Project Overview

**Project Name:** OpenCode WebChat Frontend  
**Type:** Single Page Application (SPA)  
**Framework:** Vue 3 + Vite + TypeScript  
**Purpose:** Web-based chat interface for OpenCode CLI with real-time PTY streaming via WebSocket

## 2. UI/UX Specification

### Layout Structure

```
┌─────────────────────────────────────────────┐
│  Header (56px)                              │
│  [Status] [Project Selector ▼] [Settings]  │
├─────────────────────────────────────────────┤
│                                             │
│  Main Content Area (flex: 1)                │
│  - Chat View with MessageList               │
│  - Assistant shows tool executions          │
│                                             │
├─────────────────────────────────────────────┤
│  Input Area                                 │
│  [Message input...            ] [Send]      │
├─────────────────────────────────────────────┤
│  Bottom Nav (64px)                          │
│  [Projects] [Chat] [Files] [Docs]           │
└─────────────────────────────────────────────┘
```

### Responsive Breakpoints

- **Desktop:** > 1024px (full layout with sidebar)
- **Tablet:** 768px - 1024px (collapsible sidebar)
- **Mobile:** < 768px (bottom nav only, full-screen views)

### Visual Design

**Color Palette:**
- Primary Background: `#0D1117`
- Secondary Background: `#161B22`
- Tertiary Background: `#21262D`
- Primary Accent: `#58A6FF`
- Success/Connected: `#238636`
- Text Primary: `#F0F6FC`
- Text Secondary: `#8B949E`
- Text Muted: `#484F58`
- Border: `#30363D`
- Error: `#F85149`

**Typography:**
- Font Family: `'Inter', -apple-system, BlinkMacSystemFont, sans-serif`
- Monospace: `'JetBrains Mono', 'Fira Code', monospace`
- Heading 1: 24px / 700
- Heading 2: 20px / 600
- Body: 14px / 400
- Small: 12px / 400
- Code: 13px / monospace

**Spacing System (4px base):**
- XS: 4px, SM: 8px, MD: 16px, LG: 24px, XL: 32px

### Components

#### Layout Components
- **AppHeader**: Connection status, project selector, settings
- **BottomNav**: 4-tab navigation (Projects, Chat, Files, Docs)
- **ProjectSidebar**: Project list with cards (desktop)

#### Chat Components
- **ChatView**: Main chat container
- **MessageList**: Scrollable message container
- **MessageItem**: Individual message (user/assistant)
- **InputArea**: Text input with send button
- **ToolBlock**: Tool execution display (bash commands + output)

#### Common Components
- **ConnectionStatus**: Green/red dot + device name
- **ProjectSelector**: Dropdown for project selection
- **IconButton**: Reusable icon button

## 3. Functionality Specification

### Core Features

1. **Project Management**
   - List all projects (GET /api/projects)
   - Create new project (POST /api/projects)
   - Select active project
   - Delete project (DELETE /api/projects/:id)

2. **Session Management**
   - Create new chat session per project (POST /api/sessions)
   - List sessions for project (GET /api/sessions?project_id=X)
   - Switch between sessions
   - Auto-save session on close

3. **Chat Messaging**
   - Send messages via WebSocket
   - Receive real-time responses
   - Display tool executions
   - Markdown rendering
   - Auto-scroll on new messages

4. **WebSocket Communication**
   - Connect to ws://host/ws
   - Send user input
   - Receive streaming output
   - Handle tool executions
   - Ping/pong heartbeat

### User Interactions

- Click tab to switch view
- Type message + Enter/button to send
- Click project to switch (loads sessions)
- Scroll chat history
- Click tool block to expand/collapse

## 4. Technical Specification

### Project Structure

```
web/
├── index.html
├── package.json
├── vite.config.ts
├── tsconfig.json
├── src/
│   ├── main.ts
│   ├── App.vue
│   ├── router/
│   │   └── index.ts
│   ├── stores/
│   │   ├── auth.ts
│   │   ├── projects.ts
│   │   ├── chat.ts
│   │   └── ui.ts
│   ├── services/
│   │   ├── api.ts
│   │   └── websocket.ts
│   ├── components/
│   │   ├── layout/
│   │   │   ├── AppHeader.vue
│   │   │   ├── BottomNav.vue
│   │   │   └── ProjectSidebar.vue
│   │   ├── chat/
│   │   │   ├── ChatView.vue
│   │   │   ├── MessageList.vue
│   │   │   ├── MessageItem.vue
│   │   │   ├── InputArea.vue
│   │   │   └── ToolBlock.vue
│   │   └── common/
│   │       ├── ConnectionStatus.vue
│   │       ├── ProjectSelector.vue
│   │       └── IconButton.vue
│   ├── views/
│   │   ├── ProjectsView.vue
│   │   ├── ChatView.vue
│   │   ├── FilesView.vue
│   │   └── DocsView.vue
│   ├── types/
│   │   └── index.ts
│   └── styles/
│       └── main.css
└── dist/ (build output)
```

### API Endpoints (Go Backend)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /api/projects | List all projects |
| POST | /api/projects | Create project |
| GET | /api/projects/:id | Get project |
| DELETE | /api/projects/:id | Delete project |
| GET | /api/sessions?project_id=X | List sessions |
| POST | /api/sessions | Create session |
| GET | /api/sessions/:id | Get session |
| DELETE | /api/sessions/:id | Delete session |
| GET | /api/messages?session_id=X | Get messages |
| POST | /api/messages | Create message |
| WS | /ws | WebSocket for PTY |

### WebSocket Message Format

**Outgoing:**
```json
{ "type": "input", "content": "message text" }
{ "type": "resize", "cols": 80, "rows": 24 }
```

**Incoming:**
```json
{ "type": "output", "content": "text response" }
{ "type": "tool", "name": "bash", "input": "command", "output": "result" }
{ "type": "error", "content": "error message" }
{ "type": "status", "content": "connected" }
```

### State Management (Pinia Stores)

```typescript
// useProjectStore
- projects: Project[]
- activeProject: Project | null
- fetchProjects(), createProject(), deleteProject()

// useChatStore
- sessions: Session[]
- activeSession: Session | null
- messages: Message[]
- fetchSessions(), createSession()
- fetchMessages(), sendMessage()
- wsStatus: 'connected' | 'disconnected' | 'error'

// useUIStore
- activeTab: 'projects' | 'chat' | 'files' | 'docs'
- sidebarOpen: boolean
- setActiveTab(), toggleSidebar()
```

### Dependencies

```json
{
  "vue": "^3.4.0",
  "vue-router": "^4.3.0",
  "pinia": "^2.1.0",
  "@vueuse/core": "^10.9.0",
  "axios": "^1.6.0",
  "marked": "^12.0.0",
  "highlight.js": "^11.9.0"
}
```

### TypeScript Interfaces

```typescript
interface Project {
  id: number;
  name: string;
  root_path: string;
  created_at: string;
  last_used_at: string;
}

interface Session {
  id: number;
  project_id: number;
  title: string;
  opencode_session_id: string;
  created_at: string;
  updated_at: string;
  is_active: boolean;
}

interface Message {
  id: number;
  session_id: number;
  role: 'user' | 'assistant';
  content: string;
  tool_calls: ToolCall[];
  status: 'completed' | 'interrupted' | 'error';
  created_at: string;
}

interface ToolCall {
  id: string;
  type: 'function';
  function: {
    name: string;
    arguments: string;
  };
}
```