# Tucil 3 Sliding Puzzle Solver

Project structure:
```text
tucil3-sliding-puzzle/
├── test/               # Simpan file .txt di sini
├── src/
│   ├── game/           # Gabungan Model & Engine (Logika Gerak)
│   │   ├── board.go    # Baca file & Simpan Peta
│   │   └── logic.go    # Logika sliding & validasi angka
│   ├── solver/         # Inti algoritma
│   │   ├── algorithms.go # UCS, GBFS, A* dalam satu file
│   │   └── heuristic.go  # Fungsi H1, H2, H3
│   ├── gui/            # Kode GUI (jika ambil bonus) 
│   │   └── ui.go
│   └── main.go         # Terminal menu & koordinasi
├── go.mod
└── README.md
```
