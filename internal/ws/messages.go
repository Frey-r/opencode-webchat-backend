package ws

type MessageType string

const (
	TypePrompt       MessageType = "prompt"
	TypeCancel       MessageType = "cancel"
	TypeApproveTool  MessageType = "approve_tool"
	TypeRejectTool   MessageType = "reject_tool"
	TypePing         MessageType = "ping"
	TypePong         MessageType = "pong"
	TypeToken        MessageType = "token"
	TypeToolCall     MessageType = "tool_call"
	TypeToolResult   MessageType = "tool_result"
	TypeDiffProposal MessageType = "diff_proposal"
	TypeDone         MessageType = "done"
	TypeError        MessageType = "error"
)

type InboundMessage struct {
	Type    MessageType `json:"type"`
	Content string      `json:"content,omitempty"`
	ID     string      `json:"id,omitempty"`
}

type OutboundMessage struct {
	Type    MessageType `json:"type"`
	Content string      `json:"content,omitempty"`
	ID     string      `json:"id,omitempty"`
	Data   any         `json:"data,omitempty"`
}