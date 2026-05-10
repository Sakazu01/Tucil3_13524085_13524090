# Tucil3_13524085_13524090

Program ini adalah solver Ice Sliding Puzzle berbasis pathfinding. Program membaca testcase `.txt`, menjalankan algoritma pencarian rute, lalu menampilkan hasil dan visualisasi lewat GUI web.

## Requirement

- Go 1.25 atau lebih baru
- Browser modern
- PowerShell untuk menjalankan `run.ps1` di Windows

## Compile

Program Go dapat dikompilasi dengan:

```powershell
go build ./src/backend
```

## Run

### Windows

```powershell
.\run.ps1
```

Kalau PowerShell memblokir script, jalankan dulu:

```powershell
Set-ExecutionPolicy -Scope Process Bypass
```

Kalau `run.ps1` tetap tidak bisa, jalankan manual:

```powershell
go run ./src/backend
```

### Linux / macOS

```bash
go run ./src/backend
```

Setelah server hidup, buka:

```text
http://localhost:8080
```

## Cara Pakai

Pilih testcase dari GUI atau masukkan input manual, pilih algoritma dan heuristik jika diperlukan, lalu tekan `Solve`. Hasil path, cost, waktu, iterasi, dan visualisasi board akan ditampilkan di halaman web.

## Author

- Ariel Cornelius Sitorus - 13524085
- Nashiruddin Akram - 13524090
