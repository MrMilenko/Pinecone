#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Pinecone OG Xbox id_database.json Editor (Tkinter) + MobCat Icon Cache

Edits Pinecone's current OG Xbox JSON structure:

{
  "Titles": {
    "<titleid8hex_lower>": {
      "Title Name": "Game Name",
      "Content IDs": ["<contentid16hex_lower>", ...],                 # DLC content ids
      "Title Updates": ["<tuid16hex_lower>", ...],                    # TU ids
      "Title Updates Known": [ { "<sha1>": "<tuid>:<label>", ... } ], # optional, list length 0 or 1
      "Archived": [ { "<contentid>": "<dlc name>", ... } ]            # optional, list length 0 or 1
    },
    ...
  }
}

Icon behavior:
- On title select, attempts to load icon from local cache first.
- If missing, downloads from MobCat GitHub raw and stores locally:
  https://raw.githubusercontent.com/MobCat/MobCats-original-xbox-game-list/main/icon/<TID[:4]>/<TID>.png
- Cache defaults to: <json_dir>/data/icons/
- Includes Tools:
  - Set Icon Cache Folder…
  - Prefetch Missing Icons…

Python 3 + Tkinter only. No external deps.
"""

import json
import os
import sys
import time
import hashlib
import re
import threading
import urllib.request
from dataclasses import dataclass, field
from typing import Dict, Any, Optional, List, Tuple

try:
    import tkinter as tk
    from tkinter import ttk, filedialog, messagebox
except Exception as e:
    print("Tkinter not available:", e)
    sys.exit(1)

HEX_RE = re.compile(r"^[0-9a-fA-F]+$")


# ---------------------------
# Helpers
# ---------------------------

def is_hex(s: str) -> bool:
    s = (s or "").strip()
    return bool(s) and bool(HEX_RE.match(s))

def norm_hex(s: Any, width: Optional[int] = None, lower: bool = True) -> str:
    """
    Normalize any int/'0x..'/hex string to hex, optionally width padded/truncated.
    Strips non-hex characters.
    """
    if s is None:
        return ""
    if isinstance(s, int):
        out = f"{s:X}"
    else:
        out = str(s).strip()
        if out.lower().startswith("0x"):
            out = out[2:]
    out = re.sub(r"[^0-9A-Fa-f]", "", out)
    if not out:
        return ""
    if width is not None:
        if len(out) > width:
            out = out[-width:]
        else:
            out = out.rjust(width, "0")
    out = out.lower() if lower else out.upper()
    return out

def ensure_single_dict_list(v: Any) -> List[Dict[str, str]]:
    """
    The file uses list length 0 or 1 containing a dict.
    Normalize to [] or [dict].
    """
    if v is None:
        return []
    if isinstance(v, list):
        if not v:
            return []
        if len(v) == 1 and isinstance(v[0], dict):
            d = {}
            for k, val in v[0].items():
                d[str(k)] = str(val)
            return [d]
        merged = {}
        for item in v:
            if isinstance(item, dict):
                for k, val in item.items():
                    merged[str(k)] = str(val)
        return [merged] if merged else []
    if isinstance(v, dict):
        d = {str(k): str(val) for k, val in v.items()}
        return [d] if d else []
    return []

def safe_title_obj(obj: Any) -> Dict[str, Any]:
    return obj if isinstance(obj, dict) else {}

def now_ts() -> str:
    return time.strftime("%Y%m%d-%H%M%S")

def ensure_dir(path: str) -> None:
    os.makedirs(path, exist_ok=True)


# ---------------------------
# MobCat icon fetch/cache
# ---------------------------

MOBcat_RAW_BASE = "https://raw.githubusercontent.com/MobCat/MobCats-original-xbox-game-list/main/icon"

def default_icon_cache_dir(json_path: str) -> str:
    """
    Default cache folder near the JSON (portable):
      <json_dir>/data/icons/
    """
    base = os.path.dirname(os.path.abspath(json_path))
    return os.path.join(base, "data", "icons")

def icon_cache_path(cache_dir: str, title_id_hex_8: str) -> str:
    return os.path.join(cache_dir, f"{title_id_hex_8.lower()}.png")

def mobcat_icon_url(title_id_hex_8: str) -> str:
    """
    MobCat icon path format:
      .../icon/<TID[:4]>/<TID>.png
    Example:
      .../icon/4541/4541000D.png
    """
    tid = norm_hex(title_id_hex_8, width=8, lower=False)  # upper
    prefix = tid[:4]
    return f"{MOBcat_RAW_BASE}/{prefix}/{tid}.png"

def download_to_file(url: str, out_path: str, timeout: int = 12) -> bool:
    try:
        req = urllib.request.Request(
            url,
            headers={"User-Agent": "PineconeOGXboxEditor/1.0 (+https://github.com/MrMilenko/Pinecone)"}
        )
        with urllib.request.urlopen(req, timeout=timeout) as r:
            data = r.read()

        # sanity
        if not data or len(data) < 256:
            return False
        if b"<html" in data[:512].lower():
            return False

        ensure_dir(os.path.dirname(out_path))
        with open(out_path, "wb") as f:
            f.write(data)
        return True
    except Exception:
        return False


# ---------------------------
# Model
# ---------------------------

@dataclass
class TitleRecord:
    title_id: str  # 8 hex lower
    title_name: str = ""
    content_ids: List[str] = field(default_factory=list)   # 16 hex lower
    title_updates: List[str] = field(default_factory=list) # 16 hex lower
    tu_known: Dict[str, str] = field(default_factory=dict) # sha1(lower) -> label
    archived: Dict[str, str] = field(default_factory=dict) # contentid(lower) -> name

    @staticmethod
    def from_json(title_id: str, obj: Dict[str, Any]) -> "TitleRecord":
        title_id_n = norm_hex(title_id, width=8, lower=True)
        obj = safe_title_obj(obj)

        name = str(obj.get("Title Name", "") or "")

        content_ids = []
        for cid in obj.get("Content IDs", []) or []:
            cid_n = norm_hex(cid, width=16, lower=True)
            if cid_n:
                content_ids.append(cid_n)

        title_updates = []
        for tu in obj.get("Title Updates", []) or []:
            tu_n = norm_hex(tu, width=16, lower=True)
            if tu_n:
                title_updates.append(tu_n)

        tu_known_list = ensure_single_dict_list(obj.get("Title Updates Known", []))
        tu_known = tu_known_list[0] if tu_known_list else {}

        archived_list = ensure_single_dict_list(obj.get("Archived", []))
        archived = archived_list[0] if archived_list else {}

        # normalize keys
        tu_known_norm = {}
        for sha1, label in tu_known.items():
            sha1_n = norm_hex(sha1, width=40, lower=True)
            if sha1_n:
                tu_known_norm[sha1_n] = str(label)

        archived_norm = {}
        for cid, label in archived.items():
            cid_n = norm_hex(cid, width=16, lower=True)
            if cid_n:
                archived_norm[cid_n] = str(label)

        # de-dupe / normalize lists
        content_ids = sorted(set(content_ids))
        title_updates = sorted(set(title_updates))

        return TitleRecord(
            title_id=title_id_n,
            title_name=name,
            content_ids=content_ids,
            title_updates=title_updates,
            tu_known=tu_known_norm,
            archived=archived_norm,
        )

    def to_json_obj(self) -> Dict[str, Any]:
        obj = {
            "Title Name": self.title_name or "",
            "Content IDs": sorted(set([norm_hex(x, width=16, lower=True) for x in self.content_ids if x])),
            "Title Updates": sorted(set([norm_hex(x, width=16, lower=True) for x in self.title_updates if x])),
            "Title Updates Known": [],
            "Archived": [],
        }

        # preserve the exact "list length 0 or 1" structure
        if self.tu_known:
            d = {}
            for sha1, label in self.tu_known.items():
                sha1_n = norm_hex(sha1, width=40, lower=True)
                if sha1_n:
                    d[sha1_n] = str(label)
            if d:
                obj["Title Updates Known"] = [d]

        if self.archived:
            d = {}
            for cid, label in self.archived.items():
                cid_n = norm_hex(cid, width=16, lower=True)
                if cid_n:
                    d[cid_n] = str(label)
            if d:
                obj["Archived"] = [d]

        return obj


class IdDatabaseModel:
    def __init__(self):
        self.path: Optional[str] = None
        self.dirty: bool = False
        self.titles: Dict[str, TitleRecord] = {}  # key = titleid8hex lower

    def load(self, path: str):
        with open(path, "r", encoding="utf-8") as f:
            data = json.load(f)

        if not isinstance(data, dict) or "Titles" not in data or not isinstance(data["Titles"], dict):
            raise ValueError("Not a valid OG Xbox id_database.json (missing top-level 'Titles' object)")

        titles_obj: Dict[str, Any] = data["Titles"]
        out: Dict[str, TitleRecord] = {}
        for tid, obj in titles_obj.items():
            tr = TitleRecord.from_json(tid, safe_title_obj(obj))
            if tr.title_id:
                out[tr.title_id] = tr

        self.titles = out
        self.path = path
        self.dirty = False

    def save(self, path: Optional[str] = None):
        if path is None:
            path = self.path
        if not path:
            raise ValueError("No path to save")

        # backup
        if os.path.exists(path):
            bak = f"{path}.bak-{now_ts()}"
            try:
                with open(path, "rb") as rf, open(bak, "wb") as wf:
                    wf.write(rf.read())
            except Exception as e:
                print("Backup failed:", e)

        titles_out = {}
        for tid in sorted(self.titles.keys()):
            titles_out[tid] = self.titles[tid].to_json_obj()

        out = {"Titles": titles_out}

        with open(path, "w", encoding="utf-8") as f:
            json.dump(out, f, indent=2, ensure_ascii=False)

        self.path = path
        self.dirty = False

    def mark_dirty(self):
        self.dirty = True

    def ensure_title(self, title_id: str) -> TitleRecord:
        tid = norm_hex(title_id, width=8, lower=True)
        if not tid:
            raise ValueError("TitleID must be 8 hex characters")
        if tid not in self.titles:
            self.titles[tid] = TitleRecord(title_id=tid, title_name="")
            self.mark_dirty()
        return self.titles[tid]


# ---------------------------
# UI
# ---------------------------

class IdDatabaseEditor(ttk.Frame):
    def __init__(self, master):
        super().__init__(master)
        self.master = master
        self.model = IdDatabaseModel()

        # selection state
        self._selected_title_id: Optional[str] = None

        # icon state
        self.icon_cache_dir: Optional[str] = None
        self._icon_photo: Optional[tk.PhotoImage] = None
        self._icon_mem: Dict[str, tk.PhotoImage] = {}          # tid -> PhotoImage
        self._icon_fetch_inflight: set[str] = set()            # tid
        self._placeholder_photo: Optional[tk.PhotoImage] = None

        self._build_ui()
        self._bind_accels()
        self._update_title()

    # ---- UI Build ----

    def _build_ui(self):
        self.master.title("Pinecone OG Xbox id_database.json Editor")
        self.master.geometry("1100x680")

        menubar = tk.Menu(self.master)

        filem = tk.Menu(menubar, tearoff=False)
        filem.add_command(label="Open…", accelerator="Ctrl+O", command=self.action_open)
        filem.add_command(label="Save", accelerator="Ctrl+S", command=self.action_save)
        filem.add_command(label="Save As…", accelerator="Ctrl+Shift+S", command=self.action_save_as)
        filem.add_separator()
        filem.add_command(label="Exit", command=self.action_exit)
        menubar.add_cascade(label="File", menu=filem)

        toolsm = tk.Menu(menubar, tearoff=False)
        toolsm.add_command(label="Validate", command=self.action_validate)
        toolsm.add_command(label="Compute SHA1 from file…", command=self.action_compute_sha1_global)
        toolsm.add_separator()
        toolsm.add_command(label="Set Icon Cache Folder…", command=self.action_set_icon_cache)
        toolsm.add_command(label="Prefetch Missing Icons…", command=self.action_prefetch_icons)
        menubar.add_cascade(label="Tools", menu=toolsm)

        helpm = tk.Menu(menubar, tearoff=False)
        helpm.add_command(label="About", command=self.action_about)
        menubar.add_cascade(label="Help", menu=helpm)

        self.master.config(menu=menubar)

        # layout
        self.columnconfigure(1, weight=1)
        self.rowconfigure(0, weight=1)

        left = ttk.Frame(self)
        left.grid(row=0, column=0, sticky="nsw", padx=8, pady=8)
        right = ttk.Frame(self)
        right.grid(row=0, column=1, sticky="nsew", padx=8, pady=8)
        right.columnconfigure(0, weight=1)
        right.rowconfigure(2, weight=1)

        # left: search + list + icon preview
        ttk.Label(left, text="Search Titles").grid(row=0, column=0, sticky="w")
        self.search_var = tk.StringVar()
        self.search_var.trace_add("write", lambda *_: self.populate_titles())
        ttk.Entry(left, textvariable=self.search_var, width=32).grid(row=1, column=0, sticky="we", pady=(0, 6))
        left.columnconfigure(0, weight=1)

        # icon frame
        iconf = ttk.LabelFrame(left, text="Icon")
        iconf.grid(row=2, column=0, sticky="we", pady=(0, 6))
        iconf.columnconfigure(0, weight=1)

        # placeholder: a 1x1 transparent pixel so label can hold an image
        self._placeholder_photo = tk.PhotoImage(width=1, height=1)
        self.icon_label = ttk.Label(iconf, image=self._placeholder_photo, text="(no icon)", compound="left")
        self.icon_label.grid(row=0, column=0, sticky="we", padx=6, pady=6)

        self.icon_status = ttk.Label(iconf, text="Idle")
        self.icon_status.grid(row=1, column=0, sticky="we", padx=6, pady=(0, 6))

        self.titles_list = tk.Listbox(left, height=26, exportselection=False)
        self.titles_list.grid(row=3, column=0, sticky="nswe")
        self.titles_list.bind("<<ListboxSelect>>", lambda e: self.on_title_select())
        left.rowconfigure(3, weight=1)

        # right: title header + title name editor
        header = ttk.LabelFrame(right, text="Title")
        header.grid(row=0, column=0, sticky="we")
        header.columnconfigure(1, weight=1)

        self.var_titleid = tk.StringVar()
        self.var_titlename = tk.StringVar()

        ttk.Label(header, text="TitleID:").grid(row=0, column=0, sticky="e", padx=6, pady=6)
        self.ent_titleid = ttk.Entry(header, textvariable=self.var_titleid, width=14)
        self.ent_titleid.grid(row=0, column=1, sticky="w", padx=6, pady=6)

        ttk.Button(header, text="New Title", command=self.action_new_title).grid(row=0, column=2, sticky="e", padx=6, pady=6)

        ttk.Label(header, text="Title Name:").grid(row=1, column=0, sticky="e", padx=6, pady=6)
        ttk.Entry(header, textvariable=self.var_titlename).grid(row=1, column=1, columnspan=2, sticky="we", padx=6, pady=6)

        ttk.Button(header, text="Apply Title Changes", command=self.action_apply_title).grid(row=2, column=0, columnspan=3, sticky="we", padx=6, pady=(0, 6))

        # right: notebook for DLC / TU / Known
        self.nb = ttk.Notebook(right)
        self.tab_dlc = ttk.Frame(self.nb)
        self.tab_tu = ttk.Frame(self.nb)
        self.tab_known = ttk.Frame(self.nb)
        self.nb.add(self.tab_dlc, text="DLC Content IDs")
        self.nb.add(self.tab_tu, text="Title Updates")
        self.nb.add(self.tab_known, text="Title Updates Known (SHA1)")
        self.nb.grid(row=1, column=0, sticky="we", pady=(8, 0))
        self.nb.bind("<<NotebookTabChanged>>", lambda e: self.on_tab_changed())

        # DLC tab UI
        self._build_dlc_tab(self.tab_dlc)

        # TU tab UI
        self._build_tu_tab(self.tab_tu)

        # Known tab UI
        self._build_known_tab(self.tab_known)

        # Details panel (context-sensitive)
        self.details = ttk.LabelFrame(right, text="Details")
        self.details.grid(row=2, column=0, sticky="nsew", pady=(8, 0))
        self.details.columnconfigure(1, weight=1)

        self.var_detail_id = tk.StringVar()
        self.var_detail_name = tk.StringVar()
        self.var_detail_archived = tk.BooleanVar(value=False)

        self.var_known_sha1 = tk.StringVar()
        self.var_known_value = tk.StringVar()

        ttk.Label(self.details, text="ID (hex):").grid(row=0, column=0, sticky="e", padx=6, pady=6)
        self.ent_detail_id = ttk.Entry(self.details, textvariable=self.var_detail_id)
        self.ent_detail_id.grid(row=0, column=1, sticky="we", padx=6, pady=6)

        ttk.Label(self.details, text="Name/Label:").grid(row=1, column=0, sticky="e", padx=6, pady=6)
        self.ent_detail_name = ttk.Entry(self.details, textvariable=self.var_detail_name)
        self.ent_detail_name.grid(row=1, column=1, sticky="we", padx=6, pady=6)

        self.chk_archived = ttk.Checkbutton(self.details, text="Archived (stores DLC name in Archived map)", variable=self.var_detail_archived)
        self.chk_archived.grid(row=2, column=0, columnspan=2, sticky="w", padx=6, pady=6)

        ttk.Separator(self.details, orient="horizontal").grid(row=3, column=0, columnspan=2, sticky="we", padx=6, pady=6)

        ttk.Label(self.details, text="SHA1 (40-hex):").grid(row=4, column=0, sticky="e", padx=6, pady=6)
        self.ent_known_sha1 = ttk.Entry(self.details, textvariable=self.var_known_sha1)
        self.ent_known_sha1.grid(row=4, column=1, sticky="we", padx=6, pady=6)

        ttk.Label(self.details, text="Value:").grid(row=5, column=0, sticky="e", padx=6, pady=6)
        self.ent_known_value = ttk.Entry(self.details, textvariable=self.var_known_value)
        self.ent_known_value.grid(row=5, column=1, sticky="we", padx=6, pady=6)

        btns = ttk.Frame(self.details)
        btns.grid(row=6, column=0, columnspan=2, sticky="we", padx=6, pady=(6, 6))
        ttk.Button(btns, text="Apply Changes", command=self.action_apply_entry).pack(side="left", padx=4)
        ttk.Button(btns, text="Delete Selected", command=self.action_delete_selected).pack(side="left", padx=4)
        ttk.Button(btns, text="Compute SHA1 → Known Value…", command=self.action_compute_sha1_to_known).pack(side="left", padx=4)

        self.pack(fill="both", expand=True)

        self.populate_titles()
        self._set_details_mode("DLC")

    def _build_dlc_tab(self, parent):
        parent.columnconfigure(0, weight=1)
        ttk.Label(parent, text="DLC Content IDs (from 'Content IDs'; names live in Archived map)").grid(row=0, column=0, sticky="w", padx=6, pady=6)

        self.lst_dlc = tk.Listbox(parent, height=10, exportselection=False)
        self.lst_dlc.grid(row=1, column=0, sticky="we", padx=6)
        self.lst_dlc.bind("<<ListboxSelect>>", lambda e: self.on_bucket_select("DLC"))

        btns = ttk.Frame(parent)
        btns.grid(row=2, column=0, sticky="we", padx=6, pady=6)
        ttk.Button(btns, text="Add DLC", command=self.action_add_dlc).pack(side="left", padx=3)
        ttk.Button(btns, text="Remove DLC", command=lambda: self.action_remove_from_bucket("DLC")).pack(side="left", padx=3)

    def _build_tu_tab(self, parent):
        parent.columnconfigure(0, weight=1)
        ttk.Label(parent, text="Title Updates (from 'Title Updates')").grid(row=0, column=0, sticky="w", padx=6, pady=6)

        self.lst_tu = tk.Listbox(parent, height=10, exportselection=False)
        self.lst_tu.grid(row=1, column=0, sticky="we", padx=6)
        self.lst_tu.bind("<<ListboxSelect>>", lambda e: self.on_bucket_select("TU"))

        btns = ttk.Frame(parent)
        btns.grid(row=2, column=0, sticky="we", padx=6, pady=6)
        ttk.Button(btns, text="Add TU", command=self.action_add_tu).pack(side="left", padx=3)
        ttk.Button(btns, text="Remove TU", command=lambda: self.action_remove_from_bucket("TU")).pack(side="left", padx=3)

    def _build_known_tab(self, parent):
        parent.columnconfigure(0, weight=1)
        ttk.Label(parent, text="Known Title Updates (from 'Title Updates Known'[0]: SHA1 -> value string)").grid(row=0, column=0, sticky="w", padx=6, pady=6)

        self.lst_known = tk.Listbox(parent, height=10, exportselection=False)
        self.lst_known.grid(row=1, column=0, sticky="we", padx=6)
        self.lst_known.bind("<<ListboxSelect>>", lambda e: self.on_bucket_select("KNOWN"))

        btns = ttk.Frame(parent)
        btns.grid(row=2, column=0, sticky="we", padx=6, pady=6)
        ttk.Button(btns, text="Add Known", command=self.action_add_known).pack(side="left", padx=3)
        ttk.Button(btns, text="Remove Known", command=lambda: self.action_remove_from_bucket("KNOWN")).pack(side="left", padx=3)

    def _bind_accels(self):
        self.master.bind("<Control-o>", lambda e: self.action_open())
        self.master.bind("<Control-s>", lambda e: self.action_save())
        self.master.bind("<Control-S>", lambda e: self.action_save_as())  # Shift+Ctrl+S
        self.master.protocol("WM_DELETE_WINDOW", self.action_exit)

    def _update_title(self):
        name = self.model.path or "Untitled"
        if self.model.dirty:
            name += " *"
        self.master.title(f"Pinecone OG Xbox id_database.json Editor — {name}")

    # ---- Dirty prompt ----

    def prompt_save_if_dirty(self) -> bool:
        if not self.model.dirty:
            return True
        ans = messagebox.askyesnocancel("Unsaved Changes", "Save changes before continuing?")
        if ans is None:
            return False
        if ans:
            return self.action_save()
        return True

    # ---------------------------
    # Icon logic
    # ---------------------------

    def _set_icon_status(self, txt: str) -> None:
        self.icon_status.configure(text=txt)

    def _clear_icon(self, txt: str = "(no icon)") -> None:
        self._icon_photo = self._placeholder_photo
        self.icon_label.configure(image=self._placeholder_photo, text=txt)
        self.icon_label.image = self._placeholder_photo

    def _apply_icon_photo(self, tid: str, photo: tk.PhotoImage, status: str) -> None:
        self._icon_photo = photo
        self.icon_label.configure(image=photo, text="")
        self.icon_label.image = photo
        self._set_icon_status(status)
        self._icon_mem[tid] = photo

    def _ensure_icon_cache_dir(self) -> Optional[str]:
        if self.icon_cache_dir:
            return self.icon_cache_dir
        if self.model.path:
            self.icon_cache_dir = default_icon_cache_dir(self.model.path)
            ensure_dir(self.icon_cache_dir)
            return self.icon_cache_dir
        return None

    def _load_icon_for_title(self, title_id_8: str) -> None:
        """
        Loads from memory/disk first; if missing, downloads in background then loads.
        """
        tid = norm_hex(title_id_8, width=8, lower=True)
        if not tid:
            self._clear_icon("(no title)")
            return

        cache_dir = self._ensure_icon_cache_dir()
        if not cache_dir:
            self._clear_icon("(no cache dir)")
            return

        # memory
        if tid in self._icon_mem:
            self._apply_icon_photo(tid, self._icon_mem[tid], "Icon: memory cache")
            return

        # disk
        p = icon_cache_path(cache_dir, tid)
        if os.path.exists(p):
            try:
                photo = tk.PhotoImage(file=p)
                self._apply_icon_photo(tid, photo, f"Icon: cached ({os.path.basename(p)})")
                return
            except Exception as e:
                self._set_icon_status(f"Icon decode failed, re-downloading ({e})")
                # fall through to download

        # already inflight?
        if tid in self._icon_fetch_inflight:
            self._set_icon_status("Icon: downloading…")
            return

        self._icon_fetch_inflight.add(tid)
        self._clear_icon("")

        self._set_icon_status("Icon: downloading…")

        def worker():
            ok = False
            try:
                url = mobcat_icon_url(tid)
                ok = download_to_file(url, p)
            finally:
                def done():
                    self._icon_fetch_inflight.discard(tid)
                    if ok and os.path.exists(p):
                        try:
                            photo = tk.PhotoImage(file=p)
                            self._apply_icon_photo(tid, photo, f"Icon: downloaded ({os.path.basename(p)})")
                        except Exception as e:
                            self._clear_icon("(icon failed)")
                            self._set_icon_status(f"Icon decode failed: {e}")
                    else:
                        self._clear_icon("(no icon)")
                        self._set_icon_status("Icon: not found on MobCat")
                self.master.after(0, done)

        threading.Thread(target=worker, daemon=True).start()

    def action_set_icon_cache(self):
        start = None
        if self.icon_cache_dir and os.path.isdir(self.icon_cache_dir):
            start = self.icon_cache_dir
        elif self.model.path:
            start = os.path.dirname(os.path.abspath(self.model.path))

        path = filedialog.askdirectory(title="Choose icon cache folder", initialdir=start or os.getcwd())
        if not path:
            return
        self.icon_cache_dir = path
        ensure_dir(self.icon_cache_dir)
        self._set_icon_status(f"Icon cache set: {self.icon_cache_dir}")
        # refresh current icon
        tr = self.current_title()
        if tr:
            self._load_icon_for_title(tr.title_id)

    def action_prefetch_icons(self):
        """
        Downloads missing icons for currently loaded database into cache dir.
        Runs in background; updates status line.
        """
        if not self.model.titles:
            messagebox.showinfo("Prefetch Icons", "No titles loaded.")
            return

        cache_dir = self._ensure_icon_cache_dir()
        if not cache_dir:
            messagebox.showerror("Prefetch Icons", "No cache folder set (open a JSON first or set cache folder).")
            return

        # confirm
        if not messagebox.askyesno(
            "Prefetch Missing Icons",
            f"This will download any missing icons into:\n\n{cache_dir}\n\nContinue?"
        ):
            return

        tids = sorted(self.model.titles.keys())
        self._set_icon_status("Prefetch: starting…")

        def worker():
            missing = 0
            fetched = 0
            for i, tid in enumerate(tids, start=1):
                p = icon_cache_path(cache_dir, tid)
                if os.path.exists(p):
                    continue
                missing += 1
                url = mobcat_icon_url(tid)
                ok = download_to_file(url, p)
                if ok:
                    fetched += 1
                # update occasionally
                if i % 25 == 0 or i == len(tids):
                    def upd(i=i, total=len(tids), fetched=fetched, missing=missing):
                        self._set_icon_status(f"Prefetch: {i}/{total} scanned, downloaded {fetched}/{missing} missing")
                    self.master.after(0, upd)

            def done():
                self._set_icon_status(f"Prefetch complete: downloaded {fetched}/{missing} missing")
                # refresh current selection icon
                tr = self.current_title()
                if tr:
                    self._load_icon_for_title(tr.title_id)
            self.master.after(0, done)

        threading.Thread(target=worker, daemon=True).start()

    # ---------------------------
    # Actions: File
    # ---------------------------

    def action_open(self):
        if not self.prompt_save_if_dirty():
            return
        path = filedialog.askopenfilename(
            title="Open id_database.json",
            filetypes=[("JSON files", "*.json"), ("All files", "*.*")]
        )
        if not path:
            return
        try:
            self.model.load(path)
        except Exception as e:
            messagebox.showerror("Error", f"Failed to load JSON:\n{e}")
            return

        # set default icon cache near this json unless user previously set one
        if not self.icon_cache_dir:
            self.icon_cache_dir = default_icon_cache_dir(path)
            ensure_dir(self.icon_cache_dir)

        # clear mem cache when switching databases
        self._icon_mem.clear()
        self._icon_fetch_inflight.clear()

        self.populate_titles(select_first=True)
        self._update_title()

    def action_save(self, *args):
        if not self.model.path:
            return self.action_save_as()
        try:
            self.model.save(self.model.path)
        except Exception as e:
            messagebox.showerror("Error", f"Failed to save:\n{e}")
            return False
        self._update_title()
        return True

    def action_save_as(self, *args):
        path = filedialog.asksaveasfilename(
            title="Save id_database.json As",
            defaultextension=".json",
            filetypes=[("JSON files", "*.json"), ("All files", "*.*")],
            initialfile=os.path.basename(self.model.path) if self.model.path else "id_database.json",
        )
        if not path:
            return False
        try:
            self.model.save(path)
        except Exception as e:
            messagebox.showerror("Error", f"Failed to save:\n{e}")
            return False

        # If user saved to a new location and didn't manually set cache dir, follow the json
        if self.icon_cache_dir and self.model.path:
            # keep explicit cache if user set it; otherwise, default will be used on next open.
            pass

        self._update_title()
        return True

    def action_exit(self):
        if not self.prompt_save_if_dirty():
            return
        self.master.destroy()

    def action_about(self):
        messagebox.showinfo(
            "About",
            "Pinecone OG Xbox id_database.json Editor\n"
            "Edits Pinecone's OG Xbox DLC + Title Update database.\n"
            "Also caches game icons from MobCat.\n\n"
            "Copyright © 2026 Milenko"
        )

    # ---------------------------
    # Actions: Titles
    # ---------------------------

    def action_new_title(self):
        tid = self.var_titleid.get().strip()
        if not tid:
            messagebox.showinfo("New Title", "Enter an 8-hex TitleID in the TitleID field first.")
            self.ent_titleid.focus_set()
            return
        tid_n = norm_hex(tid, width=8, lower=True)
        if len(tid_n) != 8 or not is_hex(tid_n):
            messagebox.showerror("Invalid TitleID", "TitleID must be 8 hex characters.")
            return
        if tid_n in self.model.titles:
            messagebox.showinfo("New Title", "That TitleID already exists.")
            return
        self.model.titles[tid_n] = TitleRecord(title_id=tid_n, title_name=self.var_titlename.get().strip())
        self.model.mark_dirty()
        self.populate_titles(select_title_id=tid_n)
        self._update_title()

    def action_apply_title(self):
        tr = self.current_title()
        if not tr:
            return

        new_tid_raw = self.var_titleid.get().strip()
        new_tid = norm_hex(new_tid_raw, width=8, lower=True)
        if len(new_tid) != 8 or not is_hex(new_tid):
            messagebox.showerror("Invalid TitleID", "TitleID must be 8 hex characters.")
            return

        new_name = self.var_titlename.get().strip()

        old_tid = tr.title_id
        if new_tid != old_tid:
            if new_tid in self.model.titles:
                messagebox.showerror("TitleID In Use", f"{new_tid} already exists.")
                return

            # move record
            self.model.titles[new_tid] = tr
            del self.model.titles[old_tid]
            tr.title_id = new_tid
            self._selected_title_id = new_tid

            # icon cache rename (disk)
            cache_dir = self._ensure_icon_cache_dir()
            if cache_dir:
                oldp = icon_cache_path(cache_dir, old_tid)
                newp = icon_cache_path(cache_dir, new_tid)
                try:
                    if os.path.exists(oldp) and not os.path.exists(newp):
                        os.rename(oldp, newp)
                except Exception:
                    pass

            # mem cache rename
            if old_tid in self._icon_mem and new_tid not in self._icon_mem:
                self._icon_mem[new_tid] = self._icon_mem.pop(old_tid)

        tr.title_name = new_name
        self.model.mark_dirty()
        self.populate_titles(select_title_id=tr.title_id)
        self.populate_buckets()
        self._update_title()
        self._load_icon_for_title(tr.title_id)

    # ---------------------------
    # Actions: DLC / TU / Known
    # ---------------------------

    def action_add_dlc(self):
        tr = self.current_title()
        if not tr:
            return
        base = tr.title_id
        used = set(tr.content_ids)
        cand = None
        for i in range(0, 0x100000000):
            cid = f"{base}{i:08x}"
            if cid not in used:
                cand = cid
                break
        if not cand:
            messagebox.showerror("Add DLC", "Could not find a free ContentID.")
            return
        tr.content_ids.append(cand)
        tr.content_ids = sorted(set(tr.content_ids))
        self.model.mark_dirty()
        self.populate_dlc()
        idx = tr.content_ids.index(cand)
        self.lst_dlc.selection_clear(0, "end")
        self.lst_dlc.selection_set(idx)
        self.lst_dlc.see(idx)
        self.on_bucket_select("DLC")
        self._update_title()

    def action_add_tu(self):
        tr = self.current_title()
        if not tr:
            return
        tr.title_updates.append("0000000000000000")
        self.model.mark_dirty()
        self.populate_tu()
        idx = len(tr.title_updates) - 1
        self.lst_tu.selection_clear(0, "end")
        self.lst_tu.selection_set(idx)
        self.lst_tu.see(idx)
        self.on_bucket_select("TU")
        self._update_title()

    def action_add_known(self):
        tr = self.current_title()
        if not tr:
            return
        sha = "0" * 40
        n = 0
        while sha in tr.tu_known and n < 9999:
            n += 1
            sha = (f"{n:040x}")[-40:]
        tr.tu_known[sha] = ""
        self.model.mark_dirty()
        self.populate_known()
        self.select_known_sha1(sha)
        self._update_title()

    def action_remove_from_bucket(self, bucket: str):
        tr = self.current_title()
        if not tr:
            return
        if bucket == "DLC":
            idxs = self.lst_dlc.curselection()
            if not idxs:
                return
            cid = tr.content_ids[idxs[0]]
            if not messagebox.askyesno("Remove DLC", f"Remove DLC ContentID:\n{cid}\n\nAlso removes any Archived name for it."):
                return
            tr.content_ids.pop(idxs[0])
            tr.archived.pop(cid, None)
            self.model.mark_dirty()
            self.populate_dlc()
        elif bucket == "TU":
            idxs = self.lst_tu.curselection()
            if not idxs:
                return
            tuid = tr.title_updates[idxs[0]]
            if not messagebox.askyesno("Remove TU", f"Remove Title Update ID:\n{tuid}?"):
                return
            tr.title_updates.pop(idxs[0])
            self.model.mark_dirty()
            self.populate_tu()
        elif bucket == "KNOWN":
            idxs = self.lst_known.curselection()
            if not idxs:
                return
            sha = self._known_visible_list(tr)[idxs[0]][0]
            if not messagebox.askyesno("Remove Known", f"Remove Known mapping for SHA1:\n{sha}?"):
                return
            tr.tu_known.pop(sha, None)
            self.model.mark_dirty()
            self.populate_known()

        self.clear_details()
        self._update_title()

    def action_delete_selected(self):
        tab = self._current_tab_name()
        if tab == "DLC":
            return self.action_remove_from_bucket("DLC")
        if tab == "TU":
            return self.action_remove_from_bucket("TU")
        if tab == "KNOWN":
            return self.action_remove_from_bucket("KNOWN")

    def action_apply_entry(self):
        tr = self.current_title()
        if not tr:
            return

        tab = self._current_tab_name()

        if tab in ("DLC", "TU"):
            raw_id = self.var_detail_id.get().strip()
            if tab == "DLC":
                new_id = norm_hex(raw_id, width=16, lower=True)
                if len(new_id) != 16 or not is_hex(new_id):
                    messagebox.showerror("Invalid ContentID", "ContentID must be 16 hex characters.")
                    return
                if not new_id.startswith(tr.title_id):
                    if not messagebox.askyesno(
                        "ContentID Prefix",
                        f"That ContentID does not start with TitleID {tr.title_id}.\n\nKeep it anyway?"
                    ):
                        return

                idxs = self.lst_dlc.curselection()
                if not idxs:
                    return
                old_id = tr.content_ids[idxs[0]]

                tr.content_ids[idxs[0]] = new_id
                tr.content_ids = sorted(set(tr.content_ids))

                archived = bool(self.var_detail_archived.get())
                name = self.var_detail_name.get().strip()

                if old_id != new_id:
                    if old_id in tr.archived and new_id not in tr.archived:
                        tr.archived[new_id] = tr.archived.pop(old_id)
                    else:
                        tr.archived.pop(old_id, None)

                if archived:
                    tr.archived[new_id] = name or tr.archived.get(new_id, "") or ""
                else:
                    tr.archived.pop(new_id, None)

                self.model.mark_dirty()
                self.populate_dlc(select_id=new_id)
                self._update_title()

            else:  # TU
                new_id = norm_hex(raw_id, width=16, lower=True)
                if len(new_id) != 16 or not is_hex(new_id):
                    messagebox.showerror("Invalid TU ID", "Title Update ID must be 16 hex characters.")
                    return
                idxs = self.lst_tu.curselection()
                if not idxs:
                    return
                tr.title_updates[idxs[0]] = new_id
                seen = set()
                cleaned = []
                for x in tr.title_updates:
                    if x not in seen:
                        seen.add(x)
                        cleaned.append(x)
                tr.title_updates = cleaned

                self.model.mark_dirty()
                self.populate_tu(select_id=new_id)
                self._update_title()

            return

        if tab == "KNOWN":
            sha_raw = self.var_known_sha1.get().strip()
            sha = norm_hex(sha_raw, width=40, lower=True)
            if len(sha) != 40 or not is_hex(sha):
                messagebox.showerror("Invalid SHA1", "SHA1 must be 40 hex characters.")
                return
            val = self.var_known_value.get().strip()

            idxs = self.lst_known.curselection()
            if not idxs:
                return
            old_sha = self._known_visible_list(tr)[idxs[0]][0]

            if sha != old_sha:
                if sha in tr.tu_known:
                    messagebox.showerror("SHA1 In Use", "That SHA1 key already exists.")
                    return
                tr.tu_known[sha] = tr.tu_known.pop(old_sha)

            tr.tu_known[sha] = val
            self.model.mark_dirty()
            self.populate_known(select_sha=sha)
            self._update_title()
            return

    # ---------------------------
    # Actions: Validate / SHA1 helpers
    # ---------------------------

    def action_validate(self):
        issues = []

        for tid, tr in self.model.titles.items():
            if len(tid) != 8 or not is_hex(tid):
                issues.append(f"{tid}: TitleID invalid")

            for cid in tr.content_ids:
                if len(cid) != 16 or not is_hex(cid):
                    issues.append(f"{tid} '{tr.title_name}': ContentID invalid: {cid}")
                if not cid.startswith(tid):
                    issues.append(f"{tid} '{tr.title_name}': ContentID does not start with TitleID: {cid}")

            for cid in tr.archived.keys():
                if len(cid) != 16 or not is_hex(cid):
                    issues.append(f"{tid} '{tr.title_name}': Archived ContentID invalid: {cid}")
                if cid not in tr.content_ids:
                    issues.append(f"{tid} '{tr.title_name}': Archived entry not in Content IDs: {cid}")

            for tu in tr.title_updates:
                if len(tu) != 16 or not is_hex(tu):
                    issues.append(f"{tid} '{tr.title_name}': TU invalid: {tu}")

            for sha1, val in tr.tu_known.items():
                if len(sha1) != 40 or not is_hex(sha1):
                    issues.append(f"{tid} '{tr.title_name}': Known SHA1 invalid: {sha1}")
                if val is None:
                    issues.append(f"{tid} '{tr.title_name}': Known SHA1 has empty value: {sha1}")

        if issues:
            messagebox.showwarning("Validation", "Issues found:\n\n" + "\n".join(issues[:80]) + ("\n… (more)" if len(issues) > 80 else ""))
        else:
            messagebox.showinfo("Validation", "Looks good!")

    def action_compute_sha1_global(self):
        path = filedialog.askopenfilename(title="Pick file to hash (SHA1)")
        if not path:
            return
        try:
            digest = self.compute_sha1(path)
            messagebox.showinfo("SHA1", f"SHA1 = {digest}")
        except Exception as e:
            messagebox.showerror("SHA1 Error", f"Failed to hash file:\n{e}")

    def action_compute_sha1_to_known(self):
        tab = self._current_tab_name()
        if tab != "KNOWN":
            messagebox.showinfo("SHA1 → Known", "Switch to the 'Title Updates Known (SHA1)' tab to use this button.")
            return
        path = filedialog.askopenfilename(title="Pick file to hash (SHA1)")
        if not path:
            return
        try:
            digest = self.compute_sha1(path)
            self.var_known_sha1.set(digest)
            messagebox.showinfo("SHA1", f"SHA1 = {digest}\n\nNow set the Value and click Apply Changes.")
        except Exception as e:
            messagebox.showerror("SHA1 Error", f"Failed to hash file:\n{e}")

    @staticmethod
    def compute_sha1(path: str) -> str:
        h = hashlib.sha1()
        with open(path, "rb") as f:
            for chunk in iter(lambda: f.read(1024 * 1024), b""):
                h.update(chunk)
        return h.hexdigest()

    # ---------------------------
    # Populate / Selection
    # ---------------------------

    def populate_titles(self, select_first: bool = False, select_title_id: Optional[str] = None):
        self.titles_list.delete(0, "end")

        q = (self.search_var.get() or "").lower().strip()

        visible: List[Tuple[str, str]] = []
        for tid in sorted(self.model.titles.keys()):
            tr = self.model.titles[tid]
            disp = f"{tr.title_name} ({tid})" if tr.title_name else tid
            if not q or q in disp.lower() or q in tid.lower():
                visible.append((tid, disp))
                self.titles_list.insert("end", disp)

        if not visible:
            self._selected_title_id = None
            self.clear_title_fields()
            self.populate_buckets()
            self._clear_icon("(no title)")
            self._set_icon_status("Idle")
            return

        if select_title_id:
            idx = next((i for i, (tid, _) in enumerate(visible) if tid == select_title_id), 0)
        elif self._selected_title_id:
            idx = next((i for i, (tid, _) in enumerate(visible) if tid == self._selected_title_id), 0)
        else:
            idx = 0

        if select_first:
            idx = 0

        self.titles_list.selection_set(idx)
        self.titles_list.see(idx)
        self.on_title_select()

    def on_title_select(self):
        tr = self.current_title()
        if not tr:
            self._selected_title_id = None
            self.clear_title_fields()
            self.populate_buckets()
            self._clear_icon("(no title)")
            self._set_icon_status("Idle")
            return

        self._selected_title_id = tr.title_id
        self.var_titleid.set(tr.title_id)
        self.var_titlename.set(tr.title_name)

        self.populate_buckets()
        self.clear_details()

        # icon
        self._load_icon_for_title(tr.title_id)

    def populate_buckets(self):
        self.populate_dlc()
        self.populate_tu()
        self.populate_known()

    def populate_dlc(self, select_id: Optional[str] = None):
        self.lst_dlc.delete(0, "end")
        tr = self.current_title()
        if not tr:
            return

        for cid in tr.content_ids:
            name = tr.archived.get(cid, "")
            prefix = "✓ " if cid in tr.archived else ""
            label = f"{prefix}{cid}"
            if name:
                label += f" — {name}"
            self.lst_dlc.insert("end", label)

        if tr.content_ids:
            if select_id and select_id in tr.content_ids:
                idx = tr.content_ids.index(select_id)
            else:
                idx = 0
            self.lst_dlc.selection_clear(0, "end")
            self.lst_dlc.selection_set(idx)

    def populate_tu(self, select_id: Optional[str] = None):
        self.lst_tu.delete(0, "end")
        tr = self.current_title()
        if not tr:
            return
        for tuid in tr.title_updates:
            self.lst_tu.insert("end", tuid)
        if tr.title_updates:
            idx = tr.title_updates.index(select_id) if (select_id and select_id in tr.title_updates) else 0
            self.lst_tu.selection_clear(0, "end")
            self.lst_tu.selection_set(idx)

    def _known_visible_list(self, tr: TitleRecord) -> List[Tuple[str, str]]:
        items = [(sha, tr.tu_known.get(sha, "")) for sha in tr.tu_known.keys()]
        return sorted(items, key=lambda x: ((x[1] or "").lower(), x[0]))

    def populate_known(self, select_sha: Optional[str] = None):
        self.lst_known.delete(0, "end")
        tr = self.current_title()
        if not tr:
            return
        vis = self._known_visible_list(tr)
        for sha, val in vis:
            label = f"{sha} — {val}" if val else sha
            self.lst_known.insert("end", label)
        if vis:
            if select_sha:
                idx = next((i for i, (sha, _) in enumerate(vis) if sha == select_sha), 0)
            else:
                idx = 0
            self.lst_known.selection_clear(0, "end")
            self.lst_known.selection_set(idx)

    def select_known_sha1(self, sha: str):
        tr = self.current_title()
        if not tr:
            return
        vis = self._known_visible_list(tr)
        idx = next((i for i, (s, _) in enumerate(vis) if s == sha), None)
        if idx is None:
            return
        self.lst_known.selection_clear(0, "end")
        self.lst_known.selection_set(idx)
        self.lst_known.see(idx)
        self.on_bucket_select("KNOWN")

    # ---------------------------
    # Tab / Details Mode
    # ---------------------------

    def on_tab_changed(self):
        tab = self._current_tab_name()
        self._set_details_mode(tab)
        self.on_bucket_select(tab)

    def _current_tab_name(self) -> str:
        i = self.nb.index("current")
        if i == 0:
            return "DLC"
        if i == 1:
            return "TU"
        return "KNOWN"

    def _set_details_mode(self, mode: str):
        if mode in ("DLC", "TU"):
            self.ent_detail_id.configure(state="normal")
            self.ent_detail_name.configure(state="normal")
            self.chk_archived.configure(state="normal" if mode == "DLC" else "disabled")
            self.ent_known_sha1.configure(state="disabled")
            self.ent_known_value.configure(state="disabled")
        else:
            self.ent_detail_id.configure(state="disabled")
            self.ent_detail_name.configure(state="disabled")
            self.chk_archived.configure(state="disabled")
            self.ent_known_sha1.configure(state="normal")
            self.ent_known_value.configure(state="normal")

    def on_bucket_select(self, bucket: str):
        tr = self.current_title()
        if not tr:
            return

        if bucket == "DLC":
            idxs = self.lst_dlc.curselection()
            if not idxs or idxs[0] >= len(tr.content_ids):
                return
            cid = tr.content_ids[idxs[0]]
            self.var_detail_id.set(cid)
            self.var_detail_name.set(tr.archived.get(cid, ""))
            self.var_detail_archived.set(cid in tr.archived)
            self.var_known_sha1.set("")
            self.var_known_value.set("")

        elif bucket == "TU":
            idxs = self.lst_tu.curselection()
            if not idxs or idxs[0] >= len(tr.title_updates):
                return
            tuid = tr.title_updates[idxs[0]]
            self.var_detail_id.set(tuid)
            self.var_detail_name.set("")
            self.var_detail_archived.set(False)
            self.var_known_sha1.set("")
            self.var_known_value.set("")

        elif bucket == "KNOWN":
            vis = self._known_visible_list(tr)
            idxs = self.lst_known.curselection()
            if not idxs or idxs[0] >= len(vis):
                return
            sha, val = vis[idxs[0]]
            self.var_known_sha1.set(sha)
            self.var_known_value.set(val)
            self.var_detail_id.set("")
            self.var_detail_name.set("")
            self.var_detail_archived.set(False)

    # ---------------------------
    # Helpers
    # ---------------------------

    def current_title(self) -> Optional[TitleRecord]:
        idxs = self.titles_list.curselection()
        if not idxs:
            return None

        q = (self.search_var.get() or "").lower().strip()
        visible: List[str] = []
        for tid in sorted(self.model.titles.keys()):
            tr = self.model.titles[tid]
            disp = f"{tr.title_name} ({tid})" if tr.title_name else tid
            if not q or q in disp.lower() or q in tid.lower():
                visible.append(tid)

        i = idxs[0]
        if i < 0 or i >= len(visible):
            return None
        return self.model.titles.get(visible[i])

    def clear_title_fields(self):
        self.var_titleid.set("")
        self.var_titlename.set("")

    def clear_details(self):
        self.var_detail_id.set("")
        self.var_detail_name.set("")
        self.var_detail_archived.set(False)
        self.var_known_sha1.set("")
        self.var_known_value.set("")


# ---------------------------
# main
# ---------------------------

def main():
    root = tk.Tk()
    app = IdDatabaseEditor(root)

    # optional default file if present
    default = None
    for cand in ("id_database.json",):
        if os.path.exists(cand):
            default = cand
            break
    if default:
        try:
            app.model.load(default)
            # set icon cache dir near the json by default
            app.icon_cache_dir = default_icon_cache_dir(default)
            ensure_dir(app.icon_cache_dir)

            app.populate_titles(select_first=True)
            app._update_title()
        except Exception as e:
            messagebox.showerror("Startup", f"Failed to load {default}:\n{e}")

    app.mainloop()

if __name__ == "__main__":
    main()
