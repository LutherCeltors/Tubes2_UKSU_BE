# BackEnd - Tugas Besar 2 IF2211 Strategi Algoritma  
# Tugas Besar 2 IF2211 Strategi Algoritma

## Penjelasan Singkat Algoritma

Program ini mengimplementasikan pencarian elemen pada struktur DOM-like tree menggunakan algoritma Breadth First Search (BFS) dan Depth First Search (DFS). Query yang diberikan akan diparse menjadi selector menggunakan fungsi `ParseSelector`, kemudian selector tersebut digunakan untuk menentukan apakah suatu node memenuhi kriteria pencarian.

### Depth First Search (DFS)

Implementasi DFS dilakukan secara rekursif dengan menelusuri tree dari node akar menuju daun terdalam terlebih dahulu sebelum berpindah ke cabang lain. Pendekatan ini sesuai dengan karakteristik DFS, yaitu mengeksplorasi satu jalur secara mendalam sebelum kembali ke percabangan sebelumnya.

Fungsi DFS menerima parameter `root` sebagai simpul awal, `query` sebagai selector pencarian, dan `topN` sebagai batas jumlah hasil yang ingin ditemukan. Jika `root` bernilai `nil`, fungsi akan langsung mengembalikan error.

Selama traversal, program menyimpan beberapa informasi utama:

- `results`: menyimpan node yang memenuhi kondisi selector.
- `logs`: mencatat aktivitas traversal setiap node.
- `nodesVisited`: menghitung jumlah node yang telah dikunjungi.

Pada setiap pemanggilan DFS, node akan diperiksa terlebih dahulu. Jika node bertipe `ElementNode` atau `DocumentNode`, maka node dihitung sebagai node yang dikunjungi. Jika node bertipe `ElementNode` dan memenuhi kondisi selector melalui `selector.Match(n)`, node tersebut akan ditambahkan ke dalam `results`.

Setiap node yang dikunjungi dicatat ke dalam `logs` dengan informasi `NodeID`, `Tag`, dan `Status`. Status bernilai `matched` jika node memenuhi query, dan `visited` jika node hanya dikunjungi. Traversal kemudian dilanjutkan secara rekursif ke seluruh child dari node tersebut. Jika jumlah hasil sudah mencapai batas `topN`, traversal dapat dihentikan lebih awal.

Setelah proses selesai, fungsi mengembalikan daftar node hasil pencarian, log traversal, jumlah node yang dikunjungi, serta error jika ada. Program juga menyediakan fungsi `ConvertToJSONNode` untuk mengonversi struktur `Node` menjadi `JSONNode` secara rekursif, sehingga tree dapat diserialisasi ke JSON untuk kebutuhan frontend dan visualisasi.

### Breadth First Search (BFS)

Implementasi BFS dilakukan dengan menggunakan queue untuk menyimpan node-node yang akan dikunjungi pada langkah berikutnya. BFS menelusuri tree secara melebar, yaitu memeriksa seluruh node pada level yang sama terlebih dahulu sebelum berpindah ke level berikutnya.

Pada awal pencarian, beberapa array atau struktur data disiapkan:

- `results`: menampung node yang memenuhi selector.
- `logs`: menyimpan urutan node yang telah dikunjungi beserta statusnya.
- `queue`: menampung node-node yang akan dikunjungi sesuai urutan BFS.

Ketika pencarian dimulai, simpul akar dimasukkan sebagai elemen pertama pada queue. Di setiap iterasi, elemen pertama dari queue diambil untuk diperiksa, lalu queue diperbarui agar berisi elemen berikutnya. Node kemudian dicek apakah memenuhi selector. Jika memenuhi, node dimasukkan ke dalam `results`.

Pada setiap iterasi, `logs` menyimpan `id`, `tag`, dan `status` dari node yang diperiksa. Status `matched` berarti node merupakan elemen yang dicari, sedangkan `visited` berarti node hanya dikunjungi. Setelah node diperiksa, seluruh child dari node tersebut ditambahkan ke akhir queue. Proses ini terus berjalan sampai tidak ada lagi node yang perlu diperiksa atau jumlah `results` telah mencapai batas `topN`.

## Requirement Program

Untuk menjalankan program secara manual, diperlukan:

- Go 1.25 atau versi yang kompatibel dengan `go.mod`.
- Node.js 22 atau versi yang kompatibel.
- npm.
- Browser modern untuk membuka frontend.

Untuk menjalankan program menggunakan Docker, diperlukan:

- Docker.
- Docker Compose.

## Instalasi dan Menjalankan Program

### Menjalankan Backend Secara Manual

Masuk ke folder backend:

```bash
cd Tubes2_UKSU_BE
```

Install dependency Go:

```bash
go mod download
```

Jalankan server backend:

```bash
go run .
```

Backend akan berjalan pada:

```text
http://localhost:8080
```

Jika ingin melakukan build manual di Windows:

```bash
go build -o server.exe .
.\server.exe
```

Jika ingin melakukan build manual di Linux/macOS:

```bash
go build -o server .
./server
```

### Menjalankan Frontend Secara Manual

Masuk ke folder frontend:

```bash
cd Tubes2_UKSU_FE/cauksu
```

Install dependency:

```bash
npm ci
```

Jalankan frontend:

```bash
npm run dev
```

Frontend akan berjalan pada:

```text
http://localhost:5173
```

Pada mode development, request ke `/api` akan diteruskan ke backend melalui proxy Vite dengan target default:

```text
http://localhost:8080
```

Jika ingin melakukan build frontend:

```bash
npm run build
```

Untuk menjalankan hasil build secara lokal:

```bash
npm run preview
```

### Menjalankan Backend dengan Docker

Masuk ke folder backend:

```bash
cd Tubes2_UKSU_BE
```

Build dan jalankan container backend:

```bash
docker compose up --build
```

Backend akan tersedia pada:

```text
http://localhost:8080
```

### Menjalankan Frontend dengan Docker

Pastikan backend sudah berjalan terlebih dahulu, lalu masuk ke folder frontend:

```bash
cd Tubes2_UKSU_FE/cauksu
```

Build dan jalankan container frontend:

```bash
docker compose up --build
```

Frontend akan tersedia pada:

```text
http://localhost:5173
```
## Author
- 13524021 Natanael Imandatua Manurung
- 13524052 Raynard Fausta
- 13524063 Marcel Luther Sitorus
