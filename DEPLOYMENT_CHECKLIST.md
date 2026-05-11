# 🚀 DEPLOYMENT CHECKLIST — Video Platform  
**Tanggal**: May 11, 2026  
**Status**: ✅ Backend ready to deploy

---

## 📋 STEP-BY-STEP ACTION PLAN

### ✅ COMPLETED (Backend ready)
- [x] Update config.go dengan fail-fast JWT validation
- [x] Update middleware/auth.go dengan HS256 validation
- [x] Update hls_handler.go dengan referer validation  
- [x] Update database.go untuk PostgreSQL support
- [x] Update upload_handler.go untuk Cloudinary support
- [x] Add Cloudinary SDK dependency
- [x] Backend compiles successfully (`go build`)
- [x] Backend secure for production

---

## 📝 NEXT: Prepare Credentials (10 minutes)

### 1. Get Neon DATABASE_URL
1. Go to https://neon.tech
2. Open your project dashboard
3. Copy **Connection String** (pilih tab yang punya `?sslmode=require`)
   ```
   postgresql://user:password@ep-xxx.ap-southeast-1.aws.neon.tech/dbname?sslmode=require
   ```
   > **Save this as:** `$DATABASE_URL`

### 2. Generate Secure JWT_SECRET
Run in PowerShell:
```powershell
$random = [System.Security.Cryptography.RNGCryptoServiceProvider]::new()
$bytes = [byte[]]::new(32)
$random.GetBytes($bytes)
[Convert]::ToBase64String($bytes)
```
> **Copy the output** → This is your `JWT_SECRET` (min 32 chars ✅)

### 3. Generate Secure ADMIN_PASSWORD  
```powershell
-join ((65..90) + (97..122) + (48..57) | Get-Random -Count 16 | ForEach-Object {[char]$_})
```
> **Copy the output** → This is your `ADMIN_PASSWORD`

### 4. Get Cloudinary URL
1. Go to https://cloudinary.com → Sign up
2. Dashboard → Settings → API Keys
3. Copy Cloud Name, API Key, API Secret
4. Format: `cloudinary://API_KEY:API_SECRET@CLOUD_NAME`
   > **Example:** `cloudinary://123456:abcdef@mycloud123`

---

## 🔧 NEXT: Push to GitHub (5 minutes)

```bash
# In root folder (d:\pgarcv)
cd d:\pgarcv
git add .
git commit -m "feat: ready for production deployment - PostgreSQL, Cloudinary, security hardened"
git push origin main
```

If no remote:
```bash
git remote add origin https://github.com/YOUR_USERNAME/pgarcv.git
git branch -M main
git push -u origin main
```

---

## 📤 THEN: Deploy Backend to Render (15 minutes)

### 1. Create Render Web Service
- Go to https://render.com → Sign up with GitHub
- **New** → **Web Service**
- **Connect your repository** (pgarcv)

### 2. Configure Service
```
Name:                  videoplatform-backend
Region:                Singapore
GitHub Repo:           your-repo-name
Branch:                main
Root Directory:        backend           ← IMPORTANT!
Runtime:               Go
Build Command:         go build -o server ./cmd/server
Start Command:         ./server
```

### 3. Add Environment Variables
Click **Settings** → **Environment**:

| Variable | Value |
|----------|-------|
| `PORT` | `8080` |
| `ENV` | `production` |
| `JWT_SECRET` | *paste from step 2 above* |
| `ADMIN_USERNAME` | `admin` |
| `ADMIN_PASSWORD` | *paste from step 3 above* |
| `DATABASE_URL` | *paste from step 1 above* |
| `CLOUDINARY_URL` | *paste from step 4 above* |
| `SERVER_URL` | `https://videoplatform-backend.onrender.com` |
| `UPLOAD_PATH` | `/var/uploads` |
| `HLS_OUTPUT_PATH` | `/var/hls_output` |

### 4. Deploy & Monitor
- Click **Deploy**
- Wait 2-3 minutes for build + start
- Check logs for errors
- **If success**, note the URL:
  ```
  https://videoplatform-backend.onrender.com
  ```

> ⚠️ **Cold Start**: Render free tier sleeps after 15 mins. First request takes ~30s.  
> **Fix**: Setup UptimeRobot (free) to ping backend every 5 mins.

---

## 🌐 FINALLY: Deploy Frontend to Vercel (15 minutes)

### 1. Create Vercel Project
- Go to https://vercel.com → Sign up with GitHub
- **New Project** → **Import pgarcv repository**

### 2. Configure
```
Framework:             Next.js (auto-detected)
Root Directory:        frontend
Build Command:         npm run build
Output Directory:      .next
Environment:           Node.js
Node Version:          18.x or 20.x
```

### 3. Add Environment Variable
```
NEXT_PUBLIC_API_URL = https://videoplatform-backend.onrender.com
```

### 4. Deploy  
- Click **Deploy**
- Wait 2-3 minutes
- Get URL:
  ```
  https://videoplatform-xxx.vercel.app
  ```

---

## 🔗 UPDATE Backend CORS (if needed)

If Vercel URL is different, update backend CORS:

1. Edit `backend/cmd/server/main.go` (around line 80)
2. Update `AllowOriginFunc`:
   ```go
   AllowOriginFunc: func(origin string) bool {
       allowed := []string{
           "http://localhost:3000",
           "https://videoplatform-xxx.vercel.app",  // Your Vercel URL
       }
       for _, a := range allowed {
           if origin == a {
               return true
           }
       }
       return false
   },
   ```
3. Commit & push → Render auto-redeploy

---

## ✅ FINAL CHECKLIST

```
CREDENTIALS:
[ ] Neon DATABASE_URL copied
[ ] JWT_SECRET generated (32+ chars)
[ ] ADMIN_PASSWORD generated (16 chars)
[ ] CLOUDINARY_URL copied

GITHUB:
[ ] All changes committed
[ ] Pushed to main branch

BACKEND (Render):
[ ] Web Service created
[ ] Environment variables set
[ ] Build successful
[ ] Service running
[ ] Backend URL noted

FRONTEND (Vercel):
[ ] Project created
[ ] Environment variables set
[ ] Build successful
[ ] Frontend URL noted

INTEGRATION:
[ ] Backend CORS includes Vercel URL
[ ] Can access frontend at vercel.app
[ ] Can login with admin/password
[ ] Upload works (checks Cloudinary)
[ ] Video playback works

PRODUCTION:
[ ] Monitor logs for errors
[ ] Setup UptimeRobot to prevent cold start
[ ] (Optional) Attach custom domain
```

---

## 🆘 TROUBLESHOOTING

| Problem | Solution |
|---------|----------|
| Render build fails | Check Root Directory = `backend` (not root) |
| PostgreSQL connection error | Verify `DATABASE_URL` format + `?sslmode=require` |
| Upload fails in prod | Check CLOUDINARY_URL format in Render env vars |
| CORS error in Vercel | Add Vercel URL to AllowOriginFunc, push to Render |
| Images/videos not showing | They should be in Cloudinary URLs (not `/uploads/`) |
| 502 Bad Gateway | Check Render logs, might be cold start |

---

## 💡 TIPS

- **JWT_SECRET**: Must be different between dev and prod
- **ADMIN_PASSWORD**: Never commit to Git
- **Render**: Free tier has 750 hrs/month (plenty for 1 app)
- **Neon**: 0.5GB storage free (enough for app data)
- **Cloudinary**: 25GB free (upload videos here, not Render)
- **Vercel**: Unlimited deployments, fast builds

---

**Last Updated**: May 11, 2026 | **Ready for Deploy** ✅
