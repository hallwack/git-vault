# Git Vault — Panduan Pemakaian

Git Vault adalah CLI yang mengenkripsi file secara transparan di dalam repository Git, menggunakan mekanisme clean/smudge filter bawaan Git. File tersimpan terenkripsi di Git object database (dan di remote seperti GitHub), tapi tetap terbaca normal (plaintext) di working tree lokal kamu, selama repository dalam status *unlocked*.

## Instalasi

Build dari source:

```bash
git clone https://github.com/<username>/git-vault.git
cd git-vault
go build -o git-vault ./cmd/git-vault
```

Pindahkan binary ke lokasi yang ada di `PATH`, atau jalankan langsung dari lokasi build:

```bash
./git-vault version
```

## Persiapan awal

### 1. Inisialisasi repository

Di dalam repository Git yang sudah ada (`git init` terlebih dahulu jika belum), jalankan:

```bash
git-vault init "secrets/*.env"
```

Command ini melakukan tiga hal sekaligus:

- Membuat file konfigurasi `.gitvault.yaml`.
- Mendaftarkan clean/smudge filter `git-vault` ke `.git/config` (lokal, bukan global).
- Menambahkan baris ke `.gitattributes` yang menandai pattern file mana yang harus dienkripsi.

Kalau kamu belum tahu pattern file apa yang mau dienkripsi, boleh jalankan tanpa argumen:

```bash
git-vault init
```

lalu edit `.gitattributes` secara manual nanti, contoh:

```
secrets/*.env filter=git-vault
credentials.json filter=git-vault
```

> Untuk mendaftarkan beberapa pattern sekaligus, jalankan `git-vault init` ulang dengan `--force` untuk tiap pattern, atau edit `.gitattributes` langsung — penambahan baris manual aman dan tidak akan merusak konfigurasi filter yang sudah ada.

### 2. Set password (unlock pertama kali)

```bash
git-vault unlock
```

Karena ini pertama kali, kamu akan diminta memasukkan password dua kali (konfirmasi). Dari password ini, Git Vault men-generate salt acak (disimpan di `.gitvault.salt`) dan menderivasi *master key* menggunakan Argon2id. Key ini disimpan sementara di *session* (`.git/git-vault-session`), sehingga kamu tidak perlu memasukkan password berulang kali untuk operasi Git berikutnya.

**Penting:** simpan password ini baik-baik. Git Vault tidak menyimpan password dalam bentuk apa pun — kalau lupa, file yang sudah terenkripsi tidak bisa didekripsi lagi.

## Pemakaian sehari-hari

### Encrypt file (otomatis)

Begitu repository *unlocked*, cukup pakai Git seperti biasa:

```bash
echo "API_KEY=rahasia123" > secrets/.env
git add secrets/.env
git commit -m "add secrets"
git push
```

File yang match pattern di `.gitattributes` otomatis dienkripsi oleh clean filter sebelum masuk ke Git object database. Yang ter-commit dan ter-push ke remote adalah versi terenkripsi, bukan plaintext.

### Decrypt file (otomatis)

Saat checkout, pull, atau clone repository baru:

```bash
git checkout secrets/.env
# atau setelah clone:
git clone <repo-url>
```

Smudge filter otomatis mendekripsi file kembali ke plaintext di working tree — **selama repository dalam status unlocked** dengan password yang sama.

### Lock repository

Kalau selesai kerja dan ingin menghapus master key dari cache sesi:

```bash
git-vault lock
```

Setelah ini, operasi apa pun yang butuh enkripsi/dekripsi (`git add`, `git checkout` pada file yang match pattern) akan **gagal** sampai kamu `unlock` lagi. Ini disengaja sebagai pengaman: repository yang locked tidak akan pernah diam-diam mengizinkan plaintext ter-commit.

### Unlock ulang

Untuk sesi kerja berikutnya (misal setelah restart terminal, atau di komputer lain yang sudah pernah `init` sebelumnya):

```bash
git-vault unlock
```

Kali ini kamu cukup memasukkan password sekali (tanpa konfirmasi), karena salt sudah ada.

## Cek versi

```bash
git-vault version          # versi singkat
git-vault version -v       # detail: commit, build date, versi Go, platform
```

## Kolaborasi dengan tim

Setiap kolaborator yang clone repository perlu:

1. Install `git-vault` di komputernya.
2. `git-vault unlock` dengan password yang **sama persis** yang dipakai saat setup awal (password ini perlu dibagikan lewat kanal aman di luar Git — misal password manager tim, bukan lewat chat biasa).
3. Filter clean/smudge otomatis aktif karena konfigurasi `filter.git-vault.*` dan `.gitattributes` sudah ikut ter-commit ke repository.

> `.gitvault.yaml` dan `.gitattributes` **wajib** ikut di-commit, supaya semua kolaborator pakai skema enkripsi (KDF, cipher) yang sama. Sebaliknya, `.gitvault.salt` **boleh** ikut di-commit (salt bukan rahasia) — tapi session file di `.git/git-vault-session` **tidak pernah** ikut ter-track karena memang berada di luar working tree.

## Mode non-interactive (untuk automation/CI)

Untuk skrip otomatis atau pipeline CI yang tidak punya terminal interaktif, set environment variable `GIT_VAULT_PASSWORD` sebelum memanggil `unlock`:

```bash
export GIT_VAULT_PASSWORD="password-repo-ini"
git-vault unlock
```

> Env var ini nyaman untuk automation, tapi kurang aman dibanding input interaktif (bisa terlihat proses lain di komputer yang sama lewat `/proc` di Linux). Untuk pipeline CI sungguhan, pastikan platform-nya (GitHub Actions Secrets, GitLab CI Variables, dsb) sudah mem-mask nilai secret di log sebelum dipakai dengan cara ini.

## Troubleshooting

**`git add` gagal dengan pesan "cannot encrypt: no active session"**
Repository dalam status locked. Jalankan `git-vault unlock` terlebih dahulu.

**File di working tree tidak bisa dibaca / isinya acak setelah checkout**
Kemungkinan repository locked saat checkout terjadi, atau password yang dipakai `unlock` salah. Coba `git-vault lock` lalu `git-vault unlock` ulang dengan password yang benar.

**Ingin melihat apakah filter sudah terpasang dengan benar**
```bash
cat .git/config
```
Cari section `[filter "git-vault"]` — harus ada `clean`, `smudge`, dan `required = true`.

## Batasan versi ini (MVP v0.1)

Sesuai cakupan MVP saat ini, hal-hal berikut **belum** didukung:

- Multi-user / multiple recipients (setiap orang pakai password yang sama)
- Secret scanning otomatis
- Repository audit / doctor command
- Git hooks otomatis (pre-commit, post-checkout)
- Integrasi CI/CD siap pakai (fondasi non-interactive sudah ada, integrasi platform-nya belum)
- Provider SSH, Age, atau GPG key

Lihat `roadmap.md` untuk rencana fitur-fitur ini di versi mendatang.
