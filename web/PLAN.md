# OpenCode WebChat Frontend — Implementation Plan

## Overview

Build a Vue 3 SPA that interfaces with the existing Go backend via REST API and WebSocket. The frontend provides a mobile-first chat interface for interacting with OpenCode CLI, featuring real-time streaming via WebSocket.

---

## Part 1: Tech Stack

| Category | Technology |
|----------|------------|
| Framework | Vue 3.4+ (Composition API) |
| Build Tool | Vite 5+ |
| Language | TypeScript 5+ |
| Styling | Tailwind CSS 3+ |
| Routing | Vue Router 4+ |
| State Management | Pinia 2+ |
| HTTP Client | Axios |
| WebSocket | Native WebSocket API |
| Markdown Rendering | marked + highlight.js |
| Icons | Heroicons (via @heroicons/vue) |

---

## Part 2: Project Directory Structure

```
web/
├── index.html
├── package.json
├── vite.config.ts
├── tsconfig.json
├── tailwind.config.js
├── postcss.config.js
├── src/
│   ├── main.ts                      # App entry point
│   ├── App.vue                      # Root component
│   ├── router/
│   │   └── index.ts                 # Vue Router configuration
│   ├── stores/
│   │   ├── auth.ts                  # Authentication state (JWT)
│   │   ├── projects.ts             # Projects state
│   │   ├── sessions.ts              # Chat sessions state
│   │   ├── messages.ts              # Messages state
│   │   ├── websocket.ts             # WebSocket connection state
│   │   └── ui.ts                    # UI state (active tab, sidebar)
│   ├── services/
│   │   ├── api.ts                   # Axios REST client
│   │   └── websocket.ts             # WebSocket service
│   ├── composables/
│   │   ├── useWebSocket.ts          # WebSocket composable
│   │   ├── useProjects.ts           # Projects API composable
│   │   └── useSessions.ts           # Sessions API composable
│   ├── components/
│   │   ├── layout/
│   │   │   ├── AppHeader.vue        # Top bar: menu, logo, settings
│   │   │   ├── BottomNav.vue        # Bottom tab navigation
│   │   │   └── ConnectionStatus.vue # Device/connection indicator
│   │   ├── chat/
│   │   │   ├── ChatView.vue         # Main chat container
│   │   │   ├── MessageList.vue      # Scrollable message list
│   │   │   ├── MessageItem.vue      # Individual message (user/assistant)
│   │   │   ├── InputArea.vue        # Text input with send button
│   │   │   └── ToolBlock.vue        # Tool execution display
│   │   ├── projects/
│   │   │   ├── ProjectsView.vue     # Projects list page
│   │   │   ├── ProjectCard.vue      # Project card component
│   │   │   └── InitProjectModal.vue # Create project modal
│   │   ├── docs/
│   │   │   ├── DocsView.vue         # Documentation viewer
│   │   │   ├── DocBreadcrumb.vue    # Breadcrumb navigation
│   │   │   └── MarkdownRenderer.vue # Markdown rendering
│   │   └── common/
│   │       ├── IconButton.vue       # Reusable icon button
│   │       ├── LoadingSpinner.vue   # Loading indicator
│   │       └── Modal.vue            # Modal dialog
│   ├── views/
│   │   ├── ChatPage.vue            # /chat — main chat view
│   │   ├── ProjectsPage.vue        # /projects — projects list
│   │   ├── DocsPage.vue            # /docs — documentation view
│   │   ├── LoginPage.vue           # /login — login form
│   │   └── RegisterPage.vue        # /register — registration form
│   ├── types/
│   │   └── index.ts                # TypeScript interfaces
│   └── styles/
│       └── main.css                # Tailwind imports + custom styles
└── dist/                           # Build output
```

---

## Part 3: TypeScript Interfaces

```typescript
// Project
interface Project {
  id: number;
  name: string;
  root_path: string;
  created_at: string;
  last_used_at: string;
}

// Session
interface Session {
  id: number;
  project_id: number;
  title: string;
  opencode_session_id: string;
  created_at: string;
  updated_at: string;
  is_active: boolean;
}

// Message
interface Message {
  id: number;
  session_id: number;
  role: 'user' | 'assistant';
  content: string;
  tool_calls: ToolCall[];
  status: 'completed' | 'interrupted' | 'error';
  created_at: string;
}

// Tool Call
interface ToolCall {
  id: string;
  type: 'function';
  function: {
    name: string;
    arguments: string;
  };
}

// WebSocket Messages (inbound from server)
type WSMessageType = 'token' | 'tool_call' | 'tool_result' | 'done' | 'error' | 'pong';

interface WSMessage {
  type: WSMessageType;
  content?: string;
  id?: string;
  data?: any;
}

// WebSocket Messages (outbound to server)
interface WSPromptMessage {
  type: 'prompt';
  content: string;
}

interface WSCancelMessage {
  type: 'cancel';
}

interface WSPingMessage {
  type: 'ping';
}

// UI State
type TabName = 'projects' | 'chat' | 'docs';
type WSStatus = 'connected' | 'disconnected' | 'connecting' | 'error';
```

---

## Part 4: API Client Layer

### REST API Service (`src/services/api.ts`)

Base URL: `http://localhost:8080` (configurable via `VITE_API_BASE_URL` env var)

```typescript
// Auth
POST   /api/auth/login     { username, password } → { user, token }
POST   /api/auth/logout
GET    /api/auth/me        → User

// Projects
GET    /api/projects       → Project[]
POST   /api/projects       { name, root_path } → Project
DELETE /api/projects/:id

// Sessions
GET    /api/sessions?project_id=X  → Session[]
POST   /api/sessions       { project_id, title } → Session
GET    /api/sessions/:id
PATCH  /api/sessions/:id   { title } 
DELETE /api/sessions/:id

// Messages
GET    /api/sessions/:id/messages → Message[]
```

### WebSocket Service (`src/services/websocket.ts`)

- URL: `ws://localhost:8080/ws?session_id=X&token=Y`
- Outgoing: `{ type: "prompt", content: "..." }`, `{ type: "cancel" }`, `{ type: "ping" }`
- Incoming: `{ type: "token", content: "..." }`, `{ type: "done" }`, `{ type: "error", content: "..." }`, `{ type: "tool_call", ... }`

---

## Part 5: Pinia Stores

### `useAuthStore`
```typescript
state: {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
}
actions: login(), logout(), fetchMe()
```

### `useProjectStore`
```typescript
state: {
  projects: Project[];
  activeProject: Project | null;
  isLoading: boolean;
}
actions: fetchProjects(), createProject(), setActiveProject()
```

### `useSessionStore`
```typescript
state: {
  sessions: Session[];
  activeSession: Session | null;
  isLoading: boolean;
}
actions: fetchSessions(), createSession(), setActiveSession(), deleteSession()
```

### `useMessageStore`
```typescript
state: {
  messages: Message[];
  isLoading: boolean;
}
actions: fetchMessages(), addMessage(), clearMessages()
```

### `useWebSocketStore`
```typescript
state: {
  status: WSStatus;
  lastError: string | null;
}
actions: connect(), disconnect(), send()
```

### `useUIStore`
```typescript
state: {
  activeTab: TabName;
  sidebarOpen: boolean;
}
actions: setActiveTab(), toggleSidebar()
```

---

## Part 6: Routing Structure

```typescript
const routes = [
  { path: '/login', name: 'login', component: LoginPage },
  { path: '/register', name: 'register', component: RegisterPage },
  { 
    path: '/', 
    component: MainLayout,  // AppHeader + BottomNav wrapper
    children: [
      { path: '', redirect: '/chat' },
      { path: 'chat', name: 'chat', component: ChatPage },
      { path: 'chat/:sessionId', name: 'chat-session', component: ChatPage },
      { path: 'projects', name: 'projects', component: ProjectsPage },
      { path: 'docs', name: 'docs', component: DocsPage },
    ]
  }
]
```

Navigation: BottomNav tabs → `/projects`, `/chat`, `/docs`

---

## Part 7: Implementation Phases

### Phase 1: Project Scaffold & Configuration

#### Task 1.1: Initialize Vue 3 + Vite + TypeScript Project

**Files to create:**
- `web/package.json`
- `web/vite.config.ts`
- `web/tsconfig.json`
- `web/index.html`
- `web/tailwind.config.js`
- `web/postcss.config.js`

**Steps:**
1. Create `package.json` with all dependencies
2. Create `vite.config.ts` with Vue plugin, proxy config for API
3. Create `tsconfig.json` with Vue and TypeScript config
4. Create `tailwind.config.js` with custom colors matching design spec
5. Create `postcss.config.js` for Tailwind processing
6. Create `index.html` entry point
7. Create `src/styles/main.css` with Tailwind directives

#### Task 1.2: Create TypeScript Types

**Files to create:**
- `web/src/types/index.ts`

**Steps:**
1. Define all TypeScript interfaces (Project, Session, Message, ToolCall, WSMessage types)
2. Export all types for use across the app

---

### Phase 2: State Management (Pinia Stores)

#### Task 2.1: Create All Pinia Stores

**Files to create:**
- `web/src/stores/auth.ts`
- `web/src/stores/projects.ts`
- `web/src/stores/sessions.ts`
- `web/src/stores/messages.ts`
- `web/src/stores/websocket.ts`
- `web/src/stores/ui.ts`

**Steps:**
1. Create `auth.ts` — JWT token management, login/logout actions
2. Create `projects.ts` — fetch/create/delete projects, active project state
3. Create `sessions.ts` — fetch/create sessions per project, active session state
4. Create `messages.ts` — fetch messages for session, add/clear messages
5. Create `websocket.ts` — WebSocket connection lifecycle, send/receive actions
6. Create `ui.ts` — active tab, sidebar state

---

### Phase 3: API Client & WebSocket Service

#### Task 3.1: Create REST API Service

**Files to create:**
- `web/src/services/api.ts`

**Steps:**
1. Create Axios instance with base URL and interceptors
2. Add request interceptor for JWT token header
3. Add response interceptor for error handling
4. Implement typed API methods for all endpoints
5. Export typed API functions for stores to use

#### Task 3.2: Create WebSocket Service

**Files to create:**
- `web/src/services/websocket.ts`

**Steps:**
1. Create WebSocketManager class with connect/disconnect/send methods
2. Implement message queue for offline resilience
3. Add reconnection logic with exponential backoff
4. Implement heartbeat ping/pong
5. Export singleton instance

---

### Phase 4: Routing

#### Task 4.1: Create Vue Router Configuration

**Files to create:**
- `web/src/router/index.ts`

**Steps:**
1. Define route configuration with all views
2. Create navigation guards for auth checking
3. Create MainLayout wrapper component
4. Export router instance

---

### Phase 5: Core Components — Layout

#### Task 5.1: Create AppHeader Component

**Files to create:**
- `web/src/components/layout/AppHeader.vue`

**Steps:**
1. Create top bar with hamburger menu (left), "OpenCode" logo (center), gear icon (right)
2. Add ConnectionStatus component showing device name
3. Style with Tailwind matching mobile-first design

#### Task 5.2: Create BottomNav Component

**Files to create:**
- `web/src/components/layout/BottomNav.vue`

**Steps:**
1. Create 4-tab navigation: Projects | Sessions | Docs
2. Add icons and labels for each tab
3. Use router-link for navigation
4. Highlight active tab with blue color
5. Fixed position at bottom

#### Task 5.3: Create ConnectionStatus Component

**Files to create:**
- `web/src/components/layout/ConnectionStatus.vue`

**Steps:**
1. Display gray dot + "RPi-Zero-W Connected" text
2. Add info icon with hover tooltip
3. Use store state to show connection status

---

### Phase 6: Chat Components

#### Task 6.1: Create ChatView Container

**Files to create:**
- `web/src/components/chat/ChatView.vue`

**Steps:**
1. Combine MessageList + InputArea
2. Manage scroll-to-bottom behavior
3. Handle WebSocket message display
4. Auto-connect WebSocket on mount

#### Task 6.2: Create MessageList Component

**Files to create:**
- `web/src/components/chat/MessageList.vue`

**Steps:**
1. Create scrollable container for messages
2. Implement auto-scroll to bottom on new messages
3. Add loading state indicator
4. Handle empty state ("Start a conversation")

#### Task 6.3: Create MessageItem Component

**Files to create:**
- `web/src/components/chat/MessageItem.vue`

**Steps:**
1. Create user message (right-aligned, dark background)
2. Create assistant message (left-aligned, with avatar + name)
3. Display timestamp below messages
4. Render markdown content using marked
5. Show tool executions as ToolBlock components

#### Task 6.4: Create InputArea Component

**Files to create:**
- `web/src/components/chat/InputArea.vue`

**Steps:**
1. Create text input field (multiline support)
2. Add send button with icon
3. Handle Enter key (Shift+Enter for newline)
4. Disable input while sending
5. Show connection status indicator

#### Task 6.5: Create ToolBlock Component

**Files to create:**
- `web/src/components/chat/ToolBlock.vue`

**Steps:**
1. Display bash command with dark background
2. Show command output below
3. Support for tool_call type messages
4. Collapsible for long outputs

---

### Phase 7: Projects Components

#### Task 7.1: Create ProjectsView Page

**Files to create:**
- `web/src/components/projects/ProjectsView.vue`
- `web/src/views/ProjectsPage.vue`

**Steps:**
1. Create page with "Projects" title and subtitle
2. Add "+ INIT_PROJECT" button
3. Display list of ProjectCards
4. Handle loading and empty states

#### Task 7.2: Create ProjectCard Component

**Files to create:**
- `web/src/components/projects/ProjectCard.vue`

**Steps:**
1. Display colored left border (green/blue/yellow based on status)
2. Show project name, status badge ("Running"), session count
3. Display device info (e.g., "RPi-Zero-W")
4. Show duration "2h ago"
5. White background, rounded corners, light gray border

#### Task 7.3: Create InitProjectModal

**Files to create:**
- `web/src/components/projects/InitProjectModal.vue`

**Steps:**
1. Create modal dialog with form
2. Input fields: project name, root path
3. Create and cancel buttons
4. Handle form submission

---

### Phase 8: Documentation Components

#### Task 8.1: Create DocsView Page

**Files to create:**
- `web/src/components/docs/DocsView.vue`
- `web/src/views/DocsPage.vue`

**Steps:**
1. Create breadcrumb navigation: "Repository > project-name > README.md"
2. Display document card with rendered Markdown
3. Support code blocks (dark background)
4. Support JSON blocks (light gray background)
5. Handle text formatting (headings, paragraphs)

#### Task 8.2: Create MarkdownRenderer Component

**Files to create:**
- `web/src/components/docs/MarkdownRenderer.vue`

**Steps:**
1. Use `marked` library for markdown parsing
2. Use `highlight.js` for code syntax highlighting
3. Support GFM (GitHub Flavored Markdown)
4. Sanitize output to prevent XSS

#### Task 8.3: Create DocBreadcrumb Component

**Files to create:**
- `web/src/components/docs/DocBreadcrumb.vue`

**Steps:**
1. Display path segments as clickable links
2. Show current file name
3. Handle navigation clicks

---

### Phase 9: Auth Pages

#### Task 9.1: Create LoginPage

**Files to create:**
- `web/src/views/LoginPage.vue`

**Steps:**
1. Create login form with username/password fields
2. Add submit button
3. Handle form submission via API
4. Store JWT token on success
5. Redirect to chat on success

#### Task 9.2: Create RegisterPage

**Files to create:**
- `web/src/views/RegisterPage.vue`

**Steps:**
1. Create registration form with username/password/confirmation
2. Add submit button
3. Handle form submission via API
4. Redirect to login on success

---

### Phase 10: App Entry & Integration

#### Task 10.1: Create Main.ts Entry Point

**Files to create:**
- `web/src/main.ts`

**Steps:**
1. Import and setup Pinia
2. Import and setup Vue Router
3. Mount App component

#### Task 10.2: Create App.vue Root Component

**Files to create:**
- `web/src/App.vue`

**Steps:**
1. Setup router-view for page rendering
2. Add global styles
3. Handle auth state on app load

---

## Part 8: Detailed Implementation Tasks by File

### File Creation Order

```
Phase 1:
  - web/package.json
  - web/vite.config.ts
  - web/tsconfig.json
  - web/index.html
  - web/tailwind.config.js
  - web/postcss.config.js
  - web/src/styles/main.css

Phase 2:
  - web/src/types/index.ts

Phase 3:
  - web/src/stores/auth.ts
  - web/src/stores/projects.ts
  - web/src/stores/sessions.ts
  - web/src/stores/messages.ts
  - web/src/stores/websocket.ts
  - web/src/stores/ui.ts

Phase 4:
  - web/src/services/api.ts
  - web/src/services/websocket.ts

Phase 5:
  - web/src/router/index.ts

Phase 6:
  - web/src/components/layout/AppHeader.vue
  - web/src/components/layout/BottomNav.vue
  - web/src/components/layout/ConnectionStatus.vue

Phase 7:
  - web/src/components/chat/ChatView.vue
  - web/src/components/chat/MessageList.vue
  - web/src/components/chat/MessageItem.vue
  - web/src/components/chat/InputArea.vue
  - web/src/components/chat/ToolBlock.vue

Phase 8:
  - web/src/components/projects/ProjectsView.vue
  - web/src/components/projects/ProjectCard.vue
  - web/src/components/projects/InitProjectModal.vue

Phase 9:
  - web/src/components/docs/DocsView.vue
  - web/src/components/docs/MarkdownRenderer.vue
  - web/src/components/docs/DocBreadcrumb.vue

Phase 10:
  - web/src/views/LoginPage.vue
  - web/src/views/RegisterPage.vue
  - web/src/views/ChatPage.vue
  - web/src/views/ProjectsPage.vue
  - web/src/views/DocsPage.vue

Phase 11:
  - web/src/main.ts
  - web/src/App.vue
```

---

## Part 9: Environment Configuration

### `web/.env.example`
```
VITE_API_BASE_URL=http://localhost:8080
VITE_WS_URL=ws://localhost:8080
```

### Build & Deployment

- Development: `npm run dev` — Vite dev server with HMR
- Production: `npm run build` — Output to `web/dist/`
- Backend serves `web/dist/` as static files (already configured in Go)

---

## Part 10: Testing Checklist

- [ ] Auth flow: login → store token → authenticated routes accessible
- [ ] Project creation: modal form → POST /api/projects → list updates
- [ ] Session management: select project → fetch sessions → switch sessions
- [ ] WebSocket: connect → send message → receive token stream → display
- [ ] Markdown rendering: code blocks highlighted, JSON formatted
- [ ] Responsive: mobile view with bottom nav, no horizontal overflow
- [ ] Error handling: network errors shown as toast/alert, reconnection works

---

## Part 11: Color Reference (from SPEC.md)

| Purpose | Color | Hex |
|---------|-------|-----|
| Primary Background | Dark | #0D1117 |
| Secondary Background | Dark | #161B22 |
| Tertiary Background | Dark | #21262D |
| Primary Accent | Blue | #58A6FF |
| Success/Connected | Green | #238636 |
| Text Primary | White | #F0F6FC |
| Text Secondary | Gray | #8B949E |
| Text Muted | Dark Gray | #484F58 |
| Border | Gray | #30363D |
| Error | Red | #F85149 |

---

## Part 12: Implementation Notes

### WebSocket Flow

1. User selects/creates session → `useWebSocketStore.connect(sessionId)`
2. WebSocket connects to `/ws?session_id=X`
3. User types message → `store.send({ type: 'prompt', content: '...' })`
4. Server streams tokens → `useMessageStore.addToken()` → UI updates
5. Server sends `done` → message complete, scroll to bottom

### Authentication Flow

1. User navigates to app → `useAuthStore.init()` checks for stored token
2. Token valid → fetch user profile → set authenticated
3. No token → redirect to `/login`
4. Login success → store token → redirect to `/chat`
5. Logout → clear token → redirect to `/login`

### Project/Session Flow

1. App loads → fetch projects → set first as active or show empty state
2. Select project → fetch sessions for that project
3. Select/create session → load messages for that session
4. Session is the context for WebSocket communication

---

**End of Implementation Plan**