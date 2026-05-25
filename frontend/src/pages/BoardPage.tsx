import { useState, useEffect, useRef, useCallback } from "react";
import { useAuth } from "../context/AuthContext";
import { api } from "../services/api";
import { useWebSocket } from "../hooks/useWebSocket";

interface Board {
  id: number;
  name: string;
  description: string;
  moderator_id: number;
  moderator_name: string;
  created_at: string;
}

interface TextElement {
  id: number;
  board_id: number;
  user_id: number;
  user_name: string;
  content: string;
  x: number;
  y: number;
  color: string;
  created_at: string;
  updated_at: string;
}

interface Discussion {
  id: number;
  user_name: string;
  message: string;
  created_at: string;
}

interface Member {
  id: number;
  user_id: number;
  user_name: string;
  user_email: string;
  status: string;
}

const NOTE_COLORS = [
  { bg: "#fef3c7", border: "#f59e0b", text: "#92400e" }, // yellow
  { bg: "#dbeafe", border: "#3b82f6", text: "#1e40af" }, // blue
  { bg: "#d1fae5", border: "#10b981", text: "#065f46" }, // green
  { bg: "#fce7f3", border: "#ec4899", text: "#9d174d" }, // pink
  { bg: "#e0e7ff", border: "#6366f1", text: "#3730a3" }, // indigo
  { bg: "#fed7d7", border: "#f56565", text: "#742a2a" }, // red
];

export function BoardPage() {
  const { user } = useAuth();
  const [boards, setBoards] = useState<Board[]>([]);
  const [selected, setSelected] = useState<Board | null>(null);
  const [textElements, setTextElements] = useState<TextElement[]>([]);
  const [discussions, setDiscussions] = useState<Discussion[]>([]);
  const [members, setMembers] = useState<Member[]>([]);
  const [showCreate, setShowCreate] = useState(false);
  const [newBoard, setNewBoard] = useState({ name: "", description: "" });
  const [newDiscussion, setNewDiscussion] = useState("");
  const [loading, setLoading] = useState(true);
  const [editingNote, setEditingNote] = useState<number | null>(null);
  
  // Wiki search state
  const [wikiQuery, setWikiQuery] = useState("");
  const [wikiResult, setWikiResult] = useState<{ topic: string; summary: string } | null>(null);
  const [wikiLoading, setWikiLoading] = useState(false);
  const [wikiError, setWikiError] = useState("");
  const [editContent, setEditContent] = useState("");
  const [isAddingNote, setIsAddingNote] = useState(false);
  const [draggingNote, setDraggingNote] = useState<number | null>(null);
  const [dragOffset, setDragOffset] = useState({ x: 0, y: 0 });
  const canvasRef = useRef<HTMLDivElement>(null);

  const wsRoom = selected ? `board_${selected.id}` : null;
  const { messages: wsMessages, connected: wsConnected, sendMessage: sendWsMessage } = useWebSocket(wsRoom);

  useEffect(() => {
    loadBoards();
  }, []);

  useEffect(() => {
    // Handle WebSocket messages for real-time updates
    wsMessages.forEach((msg) => {
      if (msg.type === "text_element_created") {
        setTextElements((prev) => {
          const exists = prev.find((te) => te.id === msg.data.id);
          if (exists) return prev;
          return [...prev, msg.data];
        });
      } else if (msg.type === "text_element_updated") {
        setTextElements((prev) =>
          prev.map((te) => (te.id === msg.data.id ? { ...te, ...msg.data } : te))
        );
      } else if (msg.type === "text_element_deleted") {
        setTextElements((prev) => prev.filter((te) => te.id !== msg.data.id));
      } else if (msg.type === "chat_message") {
        setDiscussions((prev) => {
          const exists = prev.find((d) => d.id === msg.data.id);
          if (exists) return prev;
          return [msg.data, ...prev];
        });
      }
    });
  }, [wsMessages]);

  const loadBoards = async () => {
    setLoading(true);
    try {
      const data = await api.listBoards();
      setBoards(data);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const selectBoard = async (b: Board) => {
    setSelected(b);
    try {
      const [elements, discs, mems] = await Promise.all([
        api.listTextElements(b.id),
        api.listDiscussions(b.id),
        api.getBoardMembers(b.id),
      ]);
      setTextElements(elements);
      setDiscussions(discs);
      setMembers(mems);
    } catch (err) {
      console.error(err);
    }
  };

  const createBoard = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await api.createBoard(newBoard);
      setShowCreate(false);
      setNewBoard({ name: "", description: "" });
      loadBoards();
    } catch (err: any) {
      alert(err.message);
    }
  };

  const createDiscussion = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selected) return;
    try {
      const disc = await api.createDiscussion({ board_id: selected.id, message: newDiscussion });
      setNewDiscussion("");
      setDiscussions((prev) => [disc, ...prev]);
      sendWsMessage("chat_message", { message: newDiscussion, discussion_id: disc.id });
    } catch (err: any) {
      alert(err.message);
    }
  };

  const joinBoard = async (boardId: number) => {
    try {
      await api.joinBoard(boardId);
      alert("Join request sent! Waiting for moderator approval.");
      loadBoards();
    } catch (err: any) {
      alert(err.message);
    }
  };

  const searchWikipedia = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!wikiQuery.trim()) return;
    
    setWikiLoading(true);
    setWikiError("");
    setWikiResult(null);
    
    try {
      const result = await api.wikiSummary(wikiQuery);
      setWikiResult(result);
    } catch (err: any) {
      setWikiError(err.message || "Failed to fetch Wikipedia data");
    } finally {
      setWikiLoading(false);
    }
  };

  const createNoteFromWiki = () => {
    if (!wikiResult || !selected) return;
    
    const canvas = canvasRef.current;
    const centerX = canvas ? canvas.clientWidth / 2 - 100 : 100;
    const centerY = canvas ? canvas.clientHeight / 2 - 75 : 100;
    
    const randomColor = NOTE_COLORS[Math.floor(Math.random() * NOTE_COLORS.length)];
    
    api.createTextElement({
      board_id: selected.id,
      content: `[${wikiResult.topic}]\n${wikiResult.summary}`,
      x: centerX,
      y: centerY,
      color: randomColor.bg,
    }).then((newElement) => {
      setTextElements((prev) => [...prev, newElement]);
      sendWsMessage("text_element_created", newElement);
    }).catch((err: any) => {
      alert(err.message);
    });
  };

  const handleCanvasClick = async (e: React.MouseEvent<HTMLDivElement>) => {
    if (!isAddingNote || !selected || !canvasRef.current) return;
    
    const rect = canvasRef.current.getBoundingClientRect();
    const x = e.clientX - rect.left;
    const y = e.clientY - rect.top;
    
    const randomColor = NOTE_COLORS[Math.floor(Math.random() * NOTE_COLORS.length)];
    
    try {
      const newElement = await api.createTextElement({
        board_id: selected.id,
        content: "New note",
        x,
        y,
        color: randomColor.bg,
      });
      setTextElements((prev) => [...prev, newElement]);
      sendWsMessage("text_element_created", newElement);
      setIsAddingNote(false);
      setEditingNote(newElement.id);
      setEditContent("New note");
    } catch (err: any) {
      alert(err.message);
    }
  };

  const updateNotePosition = async (id: number, x: number, y: number) => {
    try {
      await api.updateTextElement(id, { x, y });
      setTextElements((prev) =>
        prev.map((te) => (te.id === id ? { ...te, x, y } : te))
      );
      sendWsMessage("text_element_updated", { id, x, y });
    } catch (err) {
      console.error(err);
    }
  };

  const updateNoteContent = async (id: number) => {
    if (!editContent.trim()) return;
    try {
      await api.updateTextElement(id, { content: editContent });
      setTextElements((prev) =>
        prev.map((te) => (te.id === id ? { ...te, content: editContent } : te))
      );
      sendWsMessage("text_element_updated", { id, content: editContent });
      setEditingNote(null);
      setEditContent("");
    } catch (err) {
      console.error(err);
    }
  };

  const deleteNote = async (id: number) => {
    try {
      await api.deleteTextElement(id);
      setTextElements((prev) => prev.filter((te) => te.id !== id));
      sendWsMessage("text_element_deleted", { id });
    } catch (err) {
      console.error(err);
    }
  };

  const handleMouseDown = (e: React.MouseEvent, noteId: number) => {
    e.stopPropagation();
    const note = textElements.find((te) => te.id === noteId);
    if (!note) return;
    
    setDraggingNote(noteId);
    setDragOffset({
      x: e.clientX - note.x,
      y: e.clientY - note.y,
    });
  };

  const handleMouseMove = useCallback((e: MouseEvent) => {
    if (draggingNote === null) return;
    
    const newX = e.clientX - dragOffset.x;
    const newY = e.clientY - dragOffset.y;
    
    setTextElements((prev) =>
      prev.map((te) =>
        te.id === draggingNote ? { ...te, x: newX, y: newY } : te
      )
    );
  }, [draggingNote, dragOffset]);

  const handleMouseUp = useCallback(() => {
    if (draggingNote !== null) {
      const note = textElements.find((te) => te.id === draggingNote);
      if (note) {
        updateNotePosition(draggingNote, note.x, note.y);
      }
      setDraggingNote(null);
    }
  }, [draggingNote, textElements]);

  useEffect(() => {
    if (draggingNote !== null) {
      window.addEventListener("mousemove", handleMouseMove);
      window.addEventListener("mouseup", handleMouseUp);
      return () => {
        window.removeEventListener("mousemove", handleMouseMove);
        window.removeEventListener("mouseup", handleMouseUp);
      };
    }
  }, [draggingNote, handleMouseMove, handleMouseUp]);

  const isModerator = selected?.moderator_id === user?.id;
  const isCollaborator = user?.role === "collaborator" || user?.role === "moderator";

  if (loading) return <div style={{ padding: 24 }}>Loading...</div>;

  return (
    <div className="page">
      <div className="boardLayout">
        <aside className="sidebarStack">
          <div className="card">
            <div className="sectionTitleRow">
              <div>
                <p className="eyebrow">Browse</p>
                <h2>Boards</h2>
              </div>
              {(user?.role === "moderator" || user?.role === "superadmin") && (
                <button onClick={() => setShowCreate(true)}>New</button>
              )}
            </div>

            {showCreate && (
              <form onSubmit={createBoard} className="stack" style={{ marginTop: 18 }}>
                <input
                  placeholder="Board name"
                  value={newBoard.name}
                  onChange={(e) => setNewBoard({ ...newBoard, name: e.target.value })}
                  required
                />
                <textarea
                  placeholder="Description"
                  value={newBoard.description}
                  onChange={(e) => setNewBoard({ ...newBoard, description: e.target.value })}
                  style={{ minHeight: 100 }}
                />
                <div className="actions">
                  <button type="submit">Create</button>
                  <button type="button" className="ghost" onClick={() => setShowCreate(false)}>
                    Cancel
                  </button>
                </div>
              </form>
            )}

            <div className="boardList" style={{ marginTop: 18 }}>
              {boards.map((b) => (
                <div
                  key={b.id}
                  onClick={() => selectBoard(b)}
                  className={`boardItem${selected?.id === b.id ? " active" : ""}`}
                >
                  <div style={{ fontWeight: 600, marginBottom: 4 }}>{b.name}</div>
                  <div className="boardMeta">{b.moderator_name}</div>
                </div>
              ))}
              {boards.length === 0 && (
                <div className="emptyState surfaceBlock">
                  <p>No boards yet.</p>
                  {user?.role === "collaborator" && <p>Ask a moderator to join a board.</p>}
                </div>
              )}
            </div>
          </div>

          {selected && (
            <div className="card">
              <p className="eyebrow">Members</p>
              <div className="memberList" style={{ marginTop: 12 }}>
                {members.map((m) => (
                  <div key={m.id} className="memberItem">
                    <span className="memberName">{m.user_name}</span>
                    <span className={`memberStatus ${m.status}`}>{m.status}</span>
                  </div>
                ))}
                {members.length === 0 && (
                  <div className="emptyState">
                    <p className="muted">No members yet.</p>
                  </div>
                )}
              </div>
            </div>
          )}
        </aside>

        <section className="contentStack">
          {selected ? (
            <>
              <div className="card focusCard">
                <div className="sectionTitleRow">
                  <div>
                    <p className="eyebrow">Board</p>
                    <h1 style={{ marginBottom: 6 }}>{selected.name}</h1>
                    <p className="sub" style={{ marginBottom: 10 }}>{selected.description}</p>
                    <p className="metaText">Moderator: {selected.moderator_name}</p>
                  </div>
                  <div style={{ display: "flex", gap: 10, alignItems: "center" }}>
                    <span
                      style={{
                        fontSize: 12,
                        color: wsConnected ? "#9ce4c8" : "#ffd2cf",
                        display: "flex",
                        alignItems: "center",
                        gap: 6,
                      }}
                    >
                      <span style={{ fontSize: 10 }}>●</span> {wsConnected ? "Live" : "Offline"}
                    </span>
                    {isCollaborator && (
                      <button
                        onClick={() => setIsAddingNote(!isAddingNote)}
                        className={isAddingNote ? "active" : ""}
                      >
                        {isAddingNote ? "Click to Place" : "Add Note"}
                      </button>
                    )}
                    {user?.role === "collaborator" && (
                      <button onClick={() => joinBoard(selected.id)}>Join</button>
                    )}
                  </div>
                </div>
              </div>

              <div className="whiteboardContainer">
                <div
                  ref={canvasRef}
                  className={`whiteboardCanvas ${isAddingNote ? "adding" : ""}`}
                  onClick={handleCanvasClick}
                >
                  {textElements.map((element) => {
                    const color = NOTE_COLORS.find((c) => c.bg === element.color) || NOTE_COLORS[0];
                    const isEditing = editingNote === element.id;

                    return (
                      <div
                        key={element.id}
                        className={`stickyNote ${draggingNote === element.id ? "dragging" : ""}`}
                        style={{
                          left: element.x,
                          top: element.y,
                          backgroundColor: color.bg,
                          borderColor: color.border,
                          color: color.text,
                        }}
                        onMouseDown={(e) => !isEditing && handleMouseDown(e, element.id)}
                        onDoubleClick={() => {
                          setEditingNote(element.id);
                          setEditContent(element.content);
                        }}
                      >
                        <button
                          className="deleteNoteBtn"
                          onClick={(e) => {
                            e.stopPropagation();
                            deleteNote(element.id);
                          }}
                          title="Delete note"
                        >
                          ×
                        </button>
                        
                        {isEditing ? (
                          <textarea
                            autoFocus
                            value={editContent}
                            onChange={(e) => setEditContent(e.target.value)}
                            onBlur={() => updateNoteContent(element.id)}
                            onKeyDown={(e) => {
                              if (e.key === "Enter" && !e.shiftKey) {
                                e.preventDefault();
                                updateNoteContent(element.id);
                              }
                              if (e.key === "Escape") {
                                setEditingNote(null);
                                setEditContent("");
                              }
                            }}
                            onClick={(e) => e.stopPropagation()}
                            className="noteEditArea"
                            style={{ color: color.text }}
                          />
                        ) : (
                          <div className="noteContent">
                            <p>{element.content}</p>
                            <span className="noteAuthor">by {element.user_name}</span>
                          </div>
                        )}
                      </div>
                    );
                  })}
                  
                  {textElements.length === 0 && (
                    <div className="canvasEmptyState">
                      <p>This board is empty.</p>
                      <p className="muted">Click "Add Note" to start collaborating!</p>
                    </div>
                  )}
                </div>
              </div>

              <div className="card">
                <p className="eyebrow">Research</p>
                <form onSubmit={searchWikipedia} className="stack" style={{ marginTop: 12 }}>
                  <div style={{ display: "flex", gap: 10 }}>
                    <input
                      type="text"
                      placeholder="Search Wikipedia..."
                      value={wikiQuery}
                      onChange={(e) => setWikiQuery(e.target.value)}
                      style={{ flex: 1 }}
                    />
                    <button type="submit" disabled={wikiLoading}>
                      {wikiLoading ? "Searching..." : "Search"}
                    </button>
                  </div>
                </form>
                
                {wikiError && (
                  <div className="banner error" style={{ marginTop: 12 }}>
                    {wikiError}
                  </div>
                )}
                
                {wikiResult && (
                  <div className="wikiResult" style={{ marginTop: 16 }}>
                    <div className="wikiHeader">
                      <h4 style={{ margin: 0 }}>{wikiResult.topic}</h4>
                      <span className="wikiSource">via Wikipedia</span>
                    </div>
                    <p className="wikiContent">{wikiResult.summary}</p>
                    <div className="actions" style={{ marginTop: 12 }}>
                      <button onClick={createNoteFromWiki}>
                        Add to Board
                      </button>
                      <button className="ghost" onClick={() => setWikiResult(null)}>
                        Clear
                      </button>
                    </div>
                  </div>
                )}
                
                {!wikiResult && !wikiError && !wikiLoading && (
                  <div className="emptyState" style={{ padding: "20px 0" }}>
                    <p className="muted">Search Wikipedia to research topics and add findings to your board.</p>
                  </div>
                )}
              </div>

              <div className="card">
                <p className="eyebrow">Discussion</p>
                <form onSubmit={createDiscussion} className="stack" style={{ marginTop: 12 }}>
                  <textarea
                    placeholder="Write a message..."
                    value={newDiscussion}
                    onChange={(e) => setNewDiscussion(e.target.value)}
                    required
                    style={{ minHeight: 80 }}
                  />
                  <button type="submit">Post Message</button>
                </form>
              </div>

              <div className="discussionStack">
                {discussions.map((d) => (
                  <div key={d.id} className="surfaceBlock discussionItem">
                    <div style={{ fontWeight: 600, fontSize: 13, marginBottom: 6 }}>{d.user_name}</div>
                    <p className="metaText">{d.message}</p>
                    <p className="metaText" style={{ fontSize: 11, marginTop: 10 }}>{new Date(d.created_at).toLocaleString()}</p>
                  </div>
                ))}
                {discussions.length === 0 && (
                  <div className="emptyState surfaceBlock">No discussions yet. Be the first to post.</div>
                )}
              </div>
            </>
          ) : (
            <div className="card emptyState">
              <p>Select a board to view and collaborate.</p>
            </div>
          )}
        </section>
      </div>
    </div>
  );
}
