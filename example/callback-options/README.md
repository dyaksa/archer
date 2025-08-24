# Callback Options Example

Contoh sederhana penggunaan opsi callback pada worker Archer. Contoh ini menunjukkan:
- Callback sukses: dipanggil setelah job selesai dan status tersimpan.
- Callback gagal: dipanggil setelah job gagal (tanpa retry tersisa) dan status tersimpan.

## Prasyarat
- Go 1.20+.
- PostgreSQL berjalan lokal dan tabel `jobs` tersedia.
- Update konfigurasi koneksi di `client/main.go` dan `worker/main.go` sesuai lingkungan Anda (host, user, password, dbname, dan nama tabel jika perlu).

Schema minimal tabel `jobs` (lihat juga README utama):

```sql
CREATE TABLE jobs (
  id varchar primary key,
  queue_name varchar not null,
  status varchar not null,
  arguments jsonb not null default '{}'::jsonb,
  result jsonb not null default '{}'::jsonb,
  last_error varchar,
  retry_count integer not null default 0,
  max_retry integer not null default 0,
  retry_interval integer not null default 0,
  scheduled_at timestamptz default now(),
  started_at timestamptz,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

CREATE INDEX ON jobs (queue_name);
CREATE INDEX ON jobs (scheduled_at);
CREATE INDEX ON jobs (status);
CREATE INDEX ON jobs (started_at);
```

## Struktur
- `worker/main.go`: Mendaftarkan worker `call_test_callback` serta `WithCallbackSuccess` dan `WithCallbackFailed`.
- `client/main.go`: Menjadwalkan dua job — satu yang gagal (method `POST`) dan satu yang sukses (method `PUT`).

## Menjalankan
Jalankan dari root repo.

1) Jalankan worker
```bash
go run ./example/callback-options/worker
```

2) Pada terminal lain, jalankan client untuk menjadwalkan job
```bash
go run ./example/callback-options/client
```

Jika koneksi DB benar dan tabel tersedia, Anda akan melihat di log worker:
- Untuk job dengan method `POST` → eksekusi gagal → callback gagal dieksekusi → log "Job failed" beserta `retry_count`.
- Untuk job dengan method `PUT` → eksekusi berhasil → callback sukses dieksekusi → log "Job completed successfully".

## Catatan
- Callback dieksekusi setelah status job dipersist, sehingga tidak mempengaruhi mekanisme retry/reaper.
- Opsi callback bersifat opsional; jika tidak disetel, tidak ada efek samping.

