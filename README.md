# Backend Sharing Vision - Post Article API

REST API sederhana untuk use case **Post Article**, dibuat dengan **Golang (Gin)** dan **MySQL** untuk Test Backend Sharing Vision.

## Teknologi

- Go (Golang)
- Gin (web framework)
- MySQL (dijalankan lewat Laragon)
- golang-migrate (migration database)
- Postman (dokumentasi dan pengujian endpoint)

## Prasyarat

1. [Go](https://go.dev/dl/) sudah terinstall (cek dengan `go version`)
2. [Laragon](https://laragon.org/) (atau MySQL lain) dengan MySQL menyala di port `3306`
3. [Postman](https://www.postman.com/downloads/) (opsional, untuk mencoba endpoint)

## Cara Menjalankan

### 1. Clone repository

```bash
git clone https://github.com/USERNAME/backend-sharing-vision.git
cd backend-sharing-vision
```

### 2. Siapkan database

Buat database kosong bernama `article` (lewat phpMyAdmin atau HeidiSQL), atau jalankan:

```sql
CREATE DATABASE article;
```

Tabel `posts` **tidak perlu dibuat manual**. Tabel dibuat otomatis oleh migration saat aplikasi dijalankan.

### 3. Sesuaikan koneksi database (jika perlu)

Pengaturan bawaan ada di `main.go`:

```go
root:@tcp(127.0.0.1:3306)/article
```

Artinya user `root`, password kosong (default Laragon). Jika MySQL kamu memakai password, ubah menjadi `root:PASSWORD@tcp(127.0.0.1:3306)/article`.

### 4. Jalankan aplikasi

```bash
go mod tidy
go run main.go
```

Jika berhasil, akan muncul `Listening and serving HTTP on :8080`. API dapat diakses di `http://localhost:8080`.

Saat pertama dijalankan, migration otomatis membuat tabel `posts` dan `schema_migrations`.

## Struktur Tabel `posts`

| Kolom | Tipe | Keterangan |
|---|---|---|
| id | INT, auto increment, primary key | |
| title | VARCHAR(200) | |
| content | TEXT | |
| category | VARCHAR(100) | |
| created_date | TIMESTAMP | otomatis saat dibuat |
| updated_date | TIMESTAMP | otomatis saat diubah |
| status | VARCHAR(100) | publish / draft / thrash |

File migration ada di folder `migrations/`.

## Endpoint

| Method | URL | Deskripsi |
|---|---|---|
| POST | `/article/` | Membuat artikel baru |
| GET | `/article/:limit/:offset` | Menampilkan artikel dengan paging |
| GET | `/article/:id` | Menampilkan satu artikel berdasarkan id |
| PUT | `/article/:id` | Mengubah artikel berdasarkan id |
| DELETE | `/article/:id` | Menghapus artikel berdasarkan id |

Catatan: endpoint list mendukung filter opsional berdasarkan status, contoh `GET /article/10/0?status=publish`.

### Contoh request (POST / PUT)

```json
{
  "title": "Pengenalan REST API dengan Golang dan MySQL",
  "content": "Isi artikel minimal 200 karakter ...",
  "category": "Teknologi",
  "status": "publish"
}
```

### Contoh response

- Create: `201 Created`, body `{}`
- Get list: `200 OK`, body `[ { "id": 1, "title": "...", ... } ]`
- Get by id: `200 OK`, body `{ "id": 1, "title": "...", ... }`
- Update / Delete: `200 OK`, body `{}`
- Id tidak ditemukan: `404 Not Found`
- Validasi gagal: `400 Bad Request`, body `{ "error": "pesan kesalahan" }`

## Validasi

Validasi dijalankan sebelum artikel dibuat atau diubah:

| Field | Aturan |
|---|---|
| title | wajib, minimal 20 karakter |
| content | wajib, minimal 200 karakter |
| category | wajib, minimal 3 karakter |
| status | wajib, hanya `publish`, `draft`, atau `thrash` |

## Postman Collection

File `Sharing_Vision_Article.postman_collection.json` berisi request untuk semua endpoint beserta contoh tes validasi (folder `Create Article - Invalid`).

Cara memakai: buka Postman, klik **Import**, pilih file tersebut. Pastikan aplikasi sudah berjalan di `http://localhost:8080`.

## Struktur Project

```
backend-sharing-vision
├── migrations
│   ├── 000001_create_posts_table.up.sql
│   └── 000001_create_posts_table.down.sql
├── go.mod
├── go.sum
├── main.go
├── Sharing_Vision_Article.postman_collection.json
└── README.md
```
