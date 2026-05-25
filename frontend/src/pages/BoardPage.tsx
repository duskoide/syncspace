import { useState, useEffect } from "react";
import { useAuth } from "../context/AuthContext";
import { api } from "../services/api";

interface Board {
  id: number;
  name: string;
  description: string;
  moderator_id: number;
  moderator_name: string;
  visibility: string;
  created_at: string;
}

interface BoardImage {
  id: number;
  board_id: number;
  filename: string;
  original_name: string;
  mime_type: string;
  file_size: number;
  uploaded_by: number;
  user_name: string;
  created_at: string;
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

export function BoardPage() {
  const { user } = useAuth();
  const [boards, setBoards] = useState<Board[]>([]);
  const [selected, setSelected] = useState<Board | null>(null);
  const [images, setImages] = useState<BoardImage[]>([]);
  const [discussions, setDiscussions] = useState<Discussion[]>([]);
  const [members, setMembers] = useState<Member[]>([]);
  const [showCreate, setShowCreate] = useState(false);
  const [newBoard, setNewBoard] = useState({ name: "", description: "", visibility: "public" });
  const [newDiscussion, setNewDiscussion] = useState("");
  const [loading, setLoading] = useState(true);
  const [uploading, setUploading] = useState(false);
  
  // Wiki search state
  const [wikiQuery, setWikiQuery] = useState("");
  const [wikiResult, setWikiResult] = useState<{ topic: string; summary: string } | null>(null);
  const [wikiLoading, setWikiLoading] = useState(false);
  const [wikiError, setWikiError] = useState("");

  useEffect(() => {
    loadBoards();
  }, []);

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
      const [imgs, discs, mems] = await Promise.all([
        api.listBoardImages(b.id),
        api.listDiscussions(b.id),
        api.getBoardMembers(b.id),
      ]);
      setImages(imgs);
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
      setNewBoard({ name: "", description: "", visibility: "public" });
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
    } catch (err: any) {
      alert(err.message);
    }
  };

  const joinBoard = async (boardId: number) => {
    try {
      await api.joinBoard(boardId);
      alert("Successfully joined the board!");
      loadBoards();
      if (selected?.id === boardId) {
        selectBoard(selected);
      }
    } catch (err: any) {
      alert(err.message);
    }
  };

  const handleImageUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file || !selected) return;
    
    // Only allow images
    if (!file.type.startsWith("image/")) {
      alert("Please upload an image file");
      return;
    }
    
    setUploading(true);
    try {
      await api.uploadBoardImage(selected.id, file);
      // Refresh images
      const imgs = await api.listBoardImages(selected.id);
      setImages(imgs);
      e.target.value = ""; // Reset input
    } catch (err: any) {
      alert(err.message);
    } finally {
      setUploading(false);
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

  const isModerator = selected?.moderator_id === user?.id;
  const isCollaborator = user?.role === "collaborator" || user?.role === "moderator" || user?.role === "superadmin";

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
                <select
                  value={newBoard.visibility}
                  onChange={(e) => setNewBoard({ ...newBoard, visibility: e.target.value })}
                >
                  <option value="public">Public - Anyone can join</option>
                  <option value="private">Private - Invite only</option>
                </select>
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
                  <div style={{ fontWeight: 600, marginBottom: 4, display: "flex", justifyContent: "space-between", alignItems: "center" }}>
                    <span>{b.name}</span>
                    <span className={`tag ${b.visibility === "public" ? "tag-success" : "tag-warning"}`} style={{ fontSize: "0.7rem", padding: "2px 8px" }}>
                      {b.visibility}
                    </span>
                  </div>
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
                    <p className="metaText">
                      Moderator: {selected.moderator_name} ·{" "}
                      <span className={`tag ${selected.visibility === "public" ? "tag-success" : "tag-warning"}`} style={{ fontSize: "0.75rem" }}>
                        {selected.visibility}
                      </span>
                    </p>
                  </div>
                  <div style={{ display: "flex", gap: 10, alignItems: "center" }}>
                    {!members.find(m => m.user_id === user?.id) && user?.role !== "superadmin" && (
                      <button onClick={() => joinBoard(selected.id)}>Join Board</button>
                    )}
                  </div>
                </div>
              </div>

              {/* Image Upload Section */}
              <div className="card">
                <div className="sectionTitleRow">
                  <div>
                    <p className="eyebrow">Gallery</p>
                    <h3>Images</h3>
                  </div>
                  {isCollaborator && (
                    <div>
                      <input
                        type="file"
                        accept="image/*"
                        onChange={handleImageUpload}
                        style={{ display: "none" }}
                        id="image-upload"
                        disabled={uploading}
                      />
                      <label htmlFor="image-upload" style={{ cursor: "pointer" }}>
                        <span className="button" style={{ display: "inline-block" }}>
                          {uploading ? "Uploading..." : "Upload Image"}
                        </span>
                      </label>
                    </div>
                  )}
                </div>
                
                <div className="imageGallery" style={{ marginTop: 16 }}>
                  {images.length === 0 ? (
                    <div className="emptyState">
                      <p className="muted">No images yet. Upload some images to share!</p>
                    </div>
                  ) : (
                    <div className="imageGrid">
                      {images.map((img) => (
                        <div key={img.id} className="imageCard">
                          <img 
                            src={`${import.meta.env.VITE_API_URL || ""}/uploads/${img.filename}`}
                            alt={img.original_name}
                            className="galleryImage"
                            onError={(e) => {
                              (e.target as HTMLImageElement).src = "";
                              (e.target as HTMLImageElement).style.display = "none";
                            }}
                          />
                          <div className="imageMeta">
                            <span className="imageName" title={img.original_name}>
                              {img.original_name}
                            </span>
                            <span className="imageAuthor">by {img.user_name}</span>
                          </div>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              </div>

              {/* Wiki Research Section */}
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
                      <button className="ghost" onClick={() => setWikiResult(null)}>
                        Clear
                      </button>
                    </div>
                  </div>
                )}
                
                {!wikiResult && !wikiError && !wikiLoading && (
                  <div className="emptyState" style={{ padding: "20px 0" }}>
                    <p className="muted">Search Wikipedia to research topics for your board.</p>
                  </div>
                )}
              </div>

              {/* Discussion Section */}
              <div className="card">
                <p className="eyebrow">Discussion</p>
                {members.find(m => m.user_id === user?.id) || user?.role === "superadmin" ? (
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
                ) : (
                  <div className="emptyState">
                    <p className="muted">Join this board to participate in discussions.</p>
                  </div>
                )}
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
