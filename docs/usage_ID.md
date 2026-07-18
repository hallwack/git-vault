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

### Menambah pattern file baru

Kalau setelah `init` kamu sadar ada file/folder lain yang juga perlu dienkripsi, tidak perlu `init --force` (yang berisiko menimpa config yang sudah ada). Gunakan `track`:

```bash
git-vault track "config/credentials.json"
git-vault track "secrets/*.env" "*.pem"   # bisa lebih dari satu sekaligus
```

Command ini menambahkan pattern ke `.gitvault.yaml` **dan** `.gitattributes` sekaligus, sehingga tetap konsisten. Pattern yang sudah pernah ditambahkan tidak akan diduplikasi kalau kamu jalankan `track` lagi dengan pattern yang sama.

## Cek versi

```bash
git-vault version          # versi singkat
git-vault version -v       # detail: commit, build date, versi Go, platform
```

## Kolaborasi dengan tim

**Penting:** konfigurasi filter (`filter.git-vault.*`) disimpan di `.git/config`, yang bersifat **murni lokal** dan **tidak pernah** ikut ter-commit atau ter-clone — beda dengan `.gitattributes` yang memang bagian dari working tree. Karena itu, ada dua jalur berbeda tergantung posisi kamu:

### Kamu orang pertama yang setup (sudah dilakukan di atas)

Cukup `git-vault init` + `git-vault unlock`, seperti langkah **Persiapan awal** di atas. Setelah itu, commit dan push `.gitvault.yaml`, `.gitattributes`, dan `.gitvault.salt` ke remote, supaya kolaborator lain bisa memakainya.

### Kamu kolaborator yang clone repository yang sudah ada

**Jangan** jalankan `git-vault init` — itu akan mencoba membuat ulang konfigurasi yang sudah ada. Sebagai gantinya, gunakan `git-vault install`:

```bash
git clone <repo-url>
cd <repo>
git-vault install
git-vault unlock
```

`git-vault install` hanya mendaftarkan filter clean/smudge ke `.git/config` lokal milikmu — **tidak** membuat atau mengubah `.gitvault.yaml`, salt, maupun pattern, karena semua itu sudah ikut ter-clone. Command ini wajib dijalankan setiap kali kamu clone ulang repository di komputer baru mana pun, karena `.git/config` memang tidak pernah ikut ter-clone.

Setelah `install`, jalankan `git-vault unlock` dengan password yang **sama persis** dipakai saat setup awal — password ini perlu dibagikan lewat kanal aman di luar Git (misal password manager tim), bukan lewat chat biasa.

> **Yang wajib ikut di-commit:** `.gitvault.yaml`, `.gitattributes`, dan `.gitvault.salt`. Salt **wajib** dibagikan (bukan opsional) — karena skema saat ini berbasis password bersama (Argon2id), password yang sama hanya akan menghasilkan master key yang sama kalau salt-nya juga identik. Kalau salt hilang atau setiap kolaborator punya salt sendiri-sendiri, password yang sama pun akan menghasilkan key berbeda, dan file yang sudah terenkripsi tidak akan bisa didekripsi siapa pun. **Yang tidak boleh ikut ter-track:** session file di `.git/git-vault-session`, karena memang berada di luar working tree dan bersifat sementara per komputer.

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

## Referensi command

| Command | Kapan dipakai |
|---|---|
| `git-vault init [pattern...]` | Sekali saja, oleh orang pertama yang setup Git Vault di repository ini |
| `git-vault install` | Setiap kali clone repository yang sudah pakai Git Vault (di komputer manapun) |
| `git-vault unlock` | Awal sesi kerja, atau setelah `lock` — minta password, cache master key |
| `git-vault lock` | Akhir sesi kerja, untuk menghapus master key dari cache |
| `git-vault track <pattern...>` | Menambah pattern file baru yang perlu dienkripsi |
| `git-vault version [-v]` | Cek versi binary yang terpasang |

## Batasan versi ini (MVP v0.1)

Sesuai cakupan MVP saat ini, hal-hal berikut **belum** didukung:

- Multi-user / multiple recipients (setiap orang pakai password yang sama)
- Secret scanning otomatis
- Repository audit / doctor command
- Git hooks otomatis (pre-commit, post-checkout)
- Integrasi CI/CD siap pakai (fondasi non-interactive sudah ada, integrasi platform-nya belum)
- Provider SSH, Age, atau GPG key

Lihat `roadmap.md` untuk rencana fitur-fitur ini di versi mendatang.
