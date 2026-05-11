# 📋 PGarcv - Project Structure Documentation

**Last Updated:** May 10, 2026  
**Status:** Full-Stack Video Platform (70% Complete)

---

## 📑 Table of Contents

1. [Project Overview](#project-overview)
2. [Technology Stack](#technology-stack)
3. [Directory Structure](#directory-structure)
4. [Backend Documentation](#backend-documentation)
5. [Frontend Documentation](#frontend-documentation)
6. [Root Level Files](#root-level-files)
7. [Data Models](#data-models)
8. [API Endpoints](#api-endpoints)
9. [Environment Configuration](#environment-configuration)
10. [Development Workflow](#development-workflow)

---

## 🎯 Project Overview

**PGarcv** is a modern, full-stack video streaming and content management platform built with:

- **Purpose**: A complete video platform with public viewing and admin management interface
- **Architecture**: Monorepo with clean separation between frontend and backend
- **Deployment Ready**: Production-grade error handling, auth, and data validation
- **Status**: Core features complete (~70%), extensible for custom features (~30%)

**Key Features:**
- ✅ Public video browsing with search functionality
- ✅ Video detail page with player
- ✅ Banner slider (promotional content)
- ✅ Ad serving system (responsive to device type)
- ✅ Admin authentication (JWT + secure cookies)
- ✅ Admin dashboard with video/banner/ad management
- ✅ Dark/light theme support
- ✅ File upload system
- ✅ Responsive design (mobile-first)

---

## 🛠️ Technology Stack

### **Backend**
| Category | Technology | Version |
|----------|-----------|---------|
| Language | Go | 1.25.0 |
| Framework | Gin Web Framework | Latest |
| Database | SQLite + GORM ORM | Latest |
| Authentication | JWT (jwt-go) | Latest |
| Environment | godotenv | Latest |
| Database Migration | GORM Auto-migration | Built-in |
| Storage | Local File System | - |
| CORS | Gin CORS | Built-in |

### **Frontend**
| Category | Technology | Version |
|----------|-----------|---------|
| Framework | Next.js | 16.2.6 |
| Language | TypeScript | Latest |
| UI Library | React | 19.0.0 |
| Styling | Tailwind CSS | 4.0.0 |
| Components | Radix UI + Headless UI | Latest |
| Icons | Lucide React | Latest |
| Animation | Framer Motion | Latest |
| Theme | next-themes | Latest |
| HTTP Client | Fetch API + axios | Built-in |
| Auth Middleware | Next.js Middleware | 16.2+ |

---

## 📁 Directory Structure

```
pgarcv/
│
├── 📂 backend/                      # Go backend application
│   ├── 📂 cmd/
│   │   └── 📂 server/
│   │       └── main.go              # Entry point - Gin server setup, route definitions
│   │
│   ├── 📂 internal/                 # Private application code
│   │   ├── 📂 handler/              # HTTP request handlers (5 files)
│   │   │   ├── video_handler.go     # Video endpoints (List, Get, Create, Update, Delete)
│   │   │   ├── auth_handler.go      # Auth endpoints (Login, Logout)
│   │   │   ├── banner_handler.go    # Banner endpoints (CRUD)
│   │   │   ├── ad_handler.go        # Ad endpoints (Serve, CRUD)
│   │   │   └── upload_handler.go    # File upload processing
│   │   │
│   │   ├── 📂 middleware/           # HTTP middleware
│   │   │   └── auth.go              # JWT authentication middleware
│   │   │
│   │   ├── 📂 repository/           # Data access layer
│   │   │   ├── database.go          # GORM initialization, migrations, seed data
│   │   │   └── models.go            # Data model structs (Video, Banner, Ad)
│   │   │
│   │   └── 📂 service/              # Business logic layer (reserved for expansion)
│   │
│   ├── 📂 pkg/                      # Reusable packages
│   │   ├── 📂 response/
│   │   │   └── response.go          # API response helpers (OK, Error, Paginated)
│   │   │
│   │   └── 📂 storage/              # File storage (reserved for expansion)
│   │
│   ├── 📂 config/
│   │   └── config.go                # Configuration loader (env vars)
│   │
│   ├── 📂 migrations/               # DB migration files (GORM auto-migrate used)
│   │
│   ├── 📂 uploads/                  # Directory for uploaded media files
│   │
│   ├── 📄 go.mod                    # Go module dependencies manifest
│   ├── 📄 go.sum                    # Dependency version lock file
│   ├── 📄 .env                      # Environment variables (local development)
│   └── 📄 videoplatform.db          # SQLite database file (auto-created)
│
├── 📂 frontend/                     # Next.js frontend application
│   ├── 📂 app/                      # App Router directory structure
│   │   ├── 📄 layout.tsx            # Root layout wrapper (fonts, theme provider)
│   │   ├── 📄 globals.css           # Global styles, CSS variables (light/dark)
│   │   │
│   │   ├── 📂 (public)/             # Route group: public pages
│   │   │   ├── 📄 layout.tsx        # Public layout with navbar
│   │   │   ├── 📄 page.tsx          # Home page - video grid + banners
│   │   │   ├── 📂 watch/
│   │   │   │   └── 📂 [id]/
│   │   │   │       └── 📄 page.tsx  # Video detail page with player
│   │   │   │
│   │   │   └── 📂 search/
│   │   │       └── 📄 page.tsx      # Search results page
│   │   │
│   │   └── 📂 admin/                # Admin section
│   │       ├── 📄 login/page.tsx    # Admin login form
│   │       │
│   │       └── 📂 (authenticated)/  # Route group: protected admin pages
│   │           ├── 📄 layout.tsx    # Admin layout with sidebar
│   │           ├── 📄 dashboard/page.tsx  # Dashboard - stats & overview
│   │           │
│   │           ├── 📂 videos/
│   │           │   ├── 📄 page.tsx  # Video list management
│   │           │   ├── 📂 new/
│   │           │   │   └── 📄 page.tsx  # Create new video
│   │           │   │
│   │           │   └── 📂 edit/
│   │           │       └── 📂 [id]/
│   │           │           └── 📄 page.tsx  # Edit existing video
│   │           │
│   │           ├── 📂 banners/
│   │           │   └── 📄 page.tsx  # Banner management page
│   │           │
│   │           └── 📂 ads/
│   │               └── 📄 page.tsx  # Ad management page
│   │
│   ├── 📂 components/               # Reusable React components
│   │   ├── 📂 admin/
│   │   │   └── VideoForm.tsx        # Shared video form (create/edit)
│   │   │
│   │   ├── 📂 layout/
│   │   │   ├── Navbar.tsx           # Public navigation bar
│   │   │   └── AdminNavbar.tsx      # Admin top navigation bar
│   │   │
│   │   ├── 📂 ui/
│   │   │   ├── AdSlot.tsx           # Ad display component
│   │   │   └── ThemeToggle.tsx      # Dark/light mode toggle button
│   │   │
│   │   ├── 📂 video/
│   │   │   ├── VideoCard.tsx        # Single video thumbnail card
│   │   │   ├── VideoGridClient.tsx  # Responsive video grid layout
│   │   │   ├── BannerSlider.tsx     # Carousel banner slider
│   │   │   └── RelatedVideos.tsx    # Related videos section
│   │   │
│   │   └── 📂 providers/
│   │       └── ThemeProvider.tsx    # next-themes configuration wrapper
│   │
│   ├── 📂 lib/
│   │   └── api.ts                   # API client functions & TypeScript interfaces
│   │
│   ├── 📂 public/                   # Static assets
│   │   ├── file.svg
│   │   ├── globe.svg
│   │   ├── next.svg
│   │   ├── vercel.svg
│   │   └── window.svg
│   │
│   ├── 📄 next.config.ts            # Next.js configuration
│   ├── 📄 tsconfig.json             # TypeScript configuration
│   ├── 📄 package.json              # npm dependencies & scripts
│   ├── 📄 postcss.config.mjs        # PostCSS & Tailwind setup
│   ├── 📄 eslint.config.mjs         # ESLint rules
│   ├── 📄 next-env.d.ts             # Next.js TypeScript definitions
│   │
│   ├── 📄 .env.local                # Local environment variables
│   ├── 📄 proxy.ts                  # Next.js middleware (authentication)
│   ├── 📄 README.md                 # Frontend readme
│   ├── 📄 AGENTS.md                 # AI agent configuration
│   └── 📄 CLAUDE.md                 # Claude-specific notes
│
├── 📂 scripts/                      # Development utility scripts
│   └── reset-dev.ps1                # PowerShell script to reset dev environment
│
├── 📂 uploads/                      # Shared uploads directory (created at runtime)
│
├── 📄 EKSEKUSI_PROYEK.md           # Indonesian project execution guide
├── 📄 PROMPT_PANDUAN_PROYEK_WEB.md # Architecture guide & design decisions
├── 📄 STEP7_LENGKAP.md             # Complete code templates & snippets
└── 📄 PROJECT_STRUCTURE.md         # This file - complete structure documentation
```

---

## 🔙 Backend Documentation

### **Backend Overview**

**Location:** `backend/`  
**Language:** Go 1.25.0  
**Framework:** Gin Web Framework  
**Database:** SQLite with GORM ORM  

### **Key Files**

#### 1. **cmd/server/main.go**
- **Purpose**: Application entry point
- **Responsibilities**:
  - Initialize Gin engine
  - Setup CORS middleware
  - Register all HTTP routes
  - Start the server on configured port
  - Load environment configuration
- **Key Routes**:
  - Public: `/api/videos`, `/api/banners`, `/api/ads/serve`
  - Auth: `/api/admin/login`, `/api/admin/logout`
  - Admin (protected): `/api/admin/videos*`, `/api/admin/banners*`, `/api/admin/ads*`

#### 2. **internal/handler/video_handler.go**
- **Endpoints**:
  - `GET /api/videos` - List videos (paginated, searchable)
  - `GET /api/videos/:id` - Get single video details
  - `GET /api/videos/featured` - Featured videos for homepage
  - `POST /api/admin/videos` - Create video (protected)
  - `PUT /api/admin/videos/:id` - Update video (protected)
  - `DELETE /api/admin/videos/:id` - Delete video (protected)
  - `GET /api/admin/videos` - List all videos for admin (protected)
- **Features**:
  - Pagination support (limit, offset)
  - Search functionality (title/description)
  - Soft delete support
  - View count tracking

#### 3. **internal/handler/auth_handler.go**
- **Endpoints**:
  - `POST /api/admin/login` - Admin authentication
  - `POST /api/admin/logout` - Clear session
- **Features**:
  - JWT token generation
  - Secure cookie storage
  - HTTPS-ready (Secure flag configurable)
  - Username/password validation

#### 4. **internal/handler/banner_handler.go**
- **Endpoints**:
  - `GET /api/banners` - Get active banners (public)
  - `POST /api/admin/banners` - Create banner (protected)
  - `PUT /api/admin/banners/:id` - Update banner (protected)
  - `DELETE /api/admin/banners/:id` - Delete banner (protected)
  - `GET /api/admin/banners` - List all banners (protected)
- **Features**:
  - Sort order management
  - Active/inactive toggle
  - Image URL storage

#### 5. **internal/handler/ad_handler.go**
- **Endpoints**:
  - `GET /api/ads/serve` - Get active ads (public)
  - `POST /api/admin/ads` - Create ad (protected)
  - `PUT /api/admin/ads/:id` - Update ad (protected)
  - `DELETE /api/admin/ads/:id` - Delete ad (protected)
  - `GET /api/admin/ads` - List all ads (protected)
- **Features**:
  - Device-specific ad serving (mobile, desktop, all)
  - HTML ad code support
  - Status management

#### 6. **internal/handler/upload_handler.go**
- **Endpoints**:
  - `POST /api/admin/upload` - Upload files (protected)
- **Features**:
  - File type validation
  - Size limits
  - Secure file storage
  - UUID-based naming

#### 7. **internal/middleware/auth.go**
- **Purpose**: JWT authentication middleware
- **Implementation**:
  - Checks for JWT in cookies or Bearer token header
  - Validates token signature
  - Extracts user information
  - Returns 401 if unauthorized
- **Integration**: Applied to all `/api/admin/*` routes

#### 8. **internal/repository/database.go**
- **Responsibilities**:
  - GORM initialization with SQLite
  - Auto-migration of models
  - Seed data insertion
  - Database connection management
- **Features**:
  - Auto-creates tables on startup
  - Handles migrations automatically
  - Connection pooling

#### 9. **internal/repository/models.go**
- **Data Structures**:

  **Video Model**:
  ```go
  type Video struct {
    ID          uint      `gorm:"primaryKey"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
    Title       string
    Subtitle    string
    Description string
    ThumbnailURL string
    VideoURL    string
    PreviewURL  string
    Duration    int64       // seconds
    Views       int64
    IsPublished bool
    DeletedAt   gorm.DeletedAt
  }
  ```

  **Banner Model**:
  ```go
  type Banner struct {
    ID        uint      `gorm:"primaryKey"`
    CreatedAt time.Time
    UpdatedAt time.Time
    Title     string
    ImageURL  string
    LinkURL   string
    IsActive  bool
    SortOrder int
    DeletedAt gorm.DeletedAt
  }
  ```

  **Ad Model**:
  ```go
  type Ad struct {
    ID         uint      `gorm:"primaryKey"`
    CreatedAt  time.Time
    UpdatedAt  time.Time
    Name       string
    AdCode     string    // HTML content
    DeviceType string    // mobile, desktop, all
    IsActive   bool
    DeletedAt  gorm.DeletedAt
  }
  ```

#### 10. **config/config.go**
- **Environment Variables**:
  - `PORT` - Server port (default: 8080)
  - `JWT_SECRET` - Secret for JWT signing (required in production)
  - `DB_PATH` - SQLite database path (default: ./videoplatform.db)
  - `UPLOAD_PATH` - Upload directory (default: ./uploads)
  - `ADMIN_USERNAME` - Admin username for auth
  - `ADMIN_PASSWORD` - Admin password for auth
  - `COOKIE_SECURE` - Use secure cookies (true for HTTPS)
  - `CORS_ORIGIN` - Allowed origin for CORS (default: http://localhost:3000)

#### 11. **pkg/response/response.go**
- **Response Helpers**:
  - `OK()` - Success response with data
  - `Error()` - Error response with message
  - `Paginated()` - Paginated response with metadata

### **Backend Architecture**

```
Request → Middleware (CORS, Auth) → Handler → Repository → Database
    ↓           ↓                        ↓          ↓
  Validation   Auth Check        Business Logic  SQL Queries
                                                    ↓
                                              Response ← Response Helper
```

---

## 🎨 Frontend Documentation

### **Frontend Overview**

**Location:** `frontend/`  
**Framework:** Next.js 16.2.6 with TypeScript  
**Styling:** Tailwind CSS 4 + CSS Variables  
**Component Library:** Radix UI + Headless UI  

### **Key Files**

#### 1. **app/layout.tsx** (Root Layout)
- **Purpose**: Root wrapper for entire application
- **Features**:
  - Imports Google Fonts (Poppins, Montserrat)
  - Wraps app with ThemeProvider (dark/light mode)
  - Sets up metadata and default HTML structure
  - Initializes global styles

#### 2. **app/globals.css** (Global Styles)
- **Purpose**: Design system and CSS variables
- **Contents**:
  - Light mode color palette (backgrounds, text, borders)
  - Dark mode color palette
  - CSS custom properties for theme switching
  - Typography defaults
  - Utility classes
- **Variables Used Throughout**: Colors like `bg-slate-50`, `text-slate-900`, etc.

#### 3. **app/(public)/layout.tsx** (Public Layout)
- **Purpose**: Wrapper for public pages
- **Children**: Home, Watch, Search pages
- **Features**:
  - Renders `Navbar` component
  - Navigation between public sections
  - Consistent header across public pages

#### 4. **app/(public)/page.tsx** (Home Page)
- **Purpose**: Landing page showing all videos
- **Features**:
  - Banner slider component (promotional content)
  - Video grid component (responsive)
  - Search functionality integration
  - Ad slots (top, middle, bottom)
  - Loading states
- **Data Source**: Fetches from `/api/videos` and `/api/banners`

#### 5. **app/(public)/watch/[id]/page.tsx** (Video Detail Page)
- **Purpose**: Individual video viewing
- **Features**:
  - Video player (HTML5)
  - Video metadata (title, description, view count)
  - Related videos section
  - Ad placement
  - Comments section (extensible)
- **Dynamic Route**: Uses `[id]` for video ID from URL
- **Data Source**: Fetches from `/api/videos/:id`

#### 6. **app/(public)/search/page.tsx** (Search Page)
- **Purpose**: Search results display
- **Features**:
  - Search query from URL params
  - Filtered video results
  - Video grid layout
  - Pagination controls
- **Data Source**: Fetches from `/api/videos?search=query`

#### 7. **app/admin/login/page.tsx** (Admin Login)
- **Purpose**: Admin authentication page
- **Features**:
  - Username/password form
  - Credential validation
  - JWT token storage
  - Redirect to dashboard on success
- **Authentication**: POST to `/api/admin/login`

#### 8. **app/admin/(authenticated)/layout.tsx** (Admin Layout)
- **Purpose**: Wrapper for protected admin pages
- **Protection**: Via `proxy.ts` middleware
- **Features**:
  - Sidebar navigation
  - AdminNavbar with logout button
  - Role-based access control

#### 9. **app/admin/(authenticated)/dashboard/page.tsx**
- **Purpose**: Admin dashboard/overview
- **Features**:
  - Statistics cards (total videos, banners, ads)
  - Recent activity
  - Quick actions
  - Overview charts (extensible)

#### 10. **app/admin/(authenticated)/videos/page.tsx**
- **Purpose**: Video management list
- **Features**:
  - Table/grid of all videos
  - Edit button (links to edit page)
  - Delete button with confirmation
  - Create new button
  - Sorting and filtering

#### 11. **app/admin/(authenticated)/videos/new/page.tsx**
- **Purpose**: Create new video
- **Features**:
  - Uses shared `VideoForm` component
  - Form submission creates video
  - Redirect to videos list on success

#### 12. **app/admin/(authenticated)/videos/edit/[id]/page.tsx**
- **Purpose**: Edit existing video
- **Features**:
  - Loads video data
  - Uses shared `VideoForm` component
  - Updates video on submission
  - Pre-fills form fields

#### 13. **app/admin/(authenticated)/banners/page.tsx**
- **Purpose**: Banner management
- **Features**:
  - List all banners
  - Create/edit/delete banners
  - Drag to reorder (extensible)
  - Enable/disable toggle

#### 14. **app/admin/(authenticated)/ads/page.tsx**
- **Purpose**: Ad management
- **Features**:
  - List all ads
  - Create/edit/delete ads
  - Device type filtering
  - Enable/disable toggle

### **Components Directory**

#### **admin/VideoForm.tsx**
- **Purpose**: Reusable video form for create/edit
- **Fields**:
  - Title, Subtitle, Description
  - Thumbnail URL, Video URL, Preview URL
  - Duration, Published status
  - Upload file button
- **Usage**: Both `/new` and `/edit/[id]` pages

#### **layout/Navbar.tsx**
- **Purpose**: Public navigation bar
- **Features**:
  - Logo/branding
  - Search input with navigation
  - Theme toggle (light/dark)
  - Navigation links (Home, Browse, etc.)
  - Mobile responsive hamburger menu

#### **layout/AdminNavbar.tsx**
- **Purpose**: Admin top navigation
- **Features**:
  - Admin breadcrumbs
  - User info display
  - Logout button
  - Quick settings access

#### **ui/AdSlot.tsx**
- **Purpose**: Ad display component
- **Features**:
  - Fetches active ads from API
  - Device-specific rendering
  - Responsive sizing
  - HTML sanitization (XSS protection)

#### **ui/ThemeToggle.tsx**
- **Purpose**: Dark/light mode toggle button
- **Features**:
  - Uses next-themes for persistence
  - System preference detection
  - Smooth transitions
  - Icon switching

#### **video/VideoCard.tsx**
- **Purpose**: Single video thumbnail card
- **Features**:
  - Video thumbnail image
  - Title and subtitle
  - View count display
  - Hover effects
  - Link to watch page
  - Loading skeleton option

#### **video/VideoGridClient.tsx**
- **Purpose**: Responsive video grid layout
- **Features**:
  - CSS Grid layout (3-4 columns on desktop, 2 on tablet, 1 on mobile)
  - Maps video array to `VideoCard` components
  - Client-side rendering (marked `'use client'`)
  - Gap/spacing customization

#### **video/BannerSlider.tsx**
- **Purpose**: Carousel slider for promotions
- **Features**:
  - Framer Motion animations
  - Auto-advance slides
  - Manual navigation buttons
  - Responsive image sizing
  - Smooth transitions

#### **video/RelatedVideos.tsx**
- **Purpose**: Show related videos on detail page
- **Features**:
  - Fetches related videos (by category/tag)
  - Horizontal scroll layout
  - Quick link to watch page
  - Extensible filtering logic

#### **providers/ThemeProvider.tsx**
- **Purpose**: Theme system wrapper
- **Features**:
  - Wraps app with next-themes
  - Light/dark mode support
  - Persists user preference
  - Enables `useTheme()` hook throughout app

### **lib/api.ts** (API Client)
- **Purpose**: Centralized API communication
- **Features**:
  - Base URL configuration from env
  - TypeScript interfaces for all models
  - Request/response helpers
  - Error handling
  - Auth token management
- **Interfaces**:
  ```typescript
  interface Video {
    id: number;
    title: string;
    subtitle: string;
    description: string;
    thumbnailUrl: string;
    videoUrl: string;
    previewUrl: string;
    duration: number;
    views: number;
    isPublished: boolean;
    createdAt: string;
  }

  interface Banner {
    id: number;
    title: string;
    imageUrl: string;
    linkUrl: string;
    isActive: boolean;
    sortOrder: number;
  }

  interface Ad {
    id: number;
    name: string;
    adCode: string;
    deviceType: string;
    isActive: boolean;
  }
  ```

### **Configuration Files**

#### **next.config.ts**
- **Image Configuration**: Allowed image domains
- **Rewrites**: Proxy requests to backend
  - `/api/*` → Backend API
  - `/uploads/*` → Backend uploads
- **Headers**: Security headers setup
- **Middleware**: Authentication checks

#### **tsconfig.json**
- **Path Alias**: `@/*` maps to `./`
- **Strict Mode**: Type checking enabled
- **JSX**: React 19 jsx new transform
- **Module Resolution**: Node modules

#### **package.json**
- **Scripts**:
  - `npm run dev` - Start dev server
  - `npm run build` - Production build
  - `npm run start` - Run production build
  - `npm run lint` - Run ESLint
- **Dependencies**: Next.js, React, Tailwind, Radix UI, etc.

#### **globals.css**
- **CSS Variables** (Light Mode):
  - `--background`: Light background color
  - `--foreground`: Light text color
  - `--card`: Light card background
  - `--primary`: Brand color
  - `--secondary`: Secondary color
  - `--destructive`: Error/delete color
  - `--muted`: Muted text color
  - Border and ring colors
- **CSS Variables** (Dark Mode):
  - Same variables with dark theme colors
  - Automatically applied via `prefers-color-scheme`

#### **postcss.config.mjs**
- **Tailwind CSS Setup**: Autoprefixer, nesting
- **Content Paths**: Scans components for class usage

#### **eslint.config.mjs**
- **Linting Rules**: TypeScript, React, Next.js best practices
- **File Patterns**: `.ts`, `.tsx`, `.js`, `.jsx`

#### **proxy.ts** (Middleware)
- **Purpose**: Authenticate admin routes
- **Logic**:
  - Check JWT token in cookies
  - Redirect to `/admin/login` if unauthorized
  - Allow access to protected admin pages if authorized
  - Bypass for login page

#### **.env.local** (Example)
```
NEXT_PUBLIC_API_URL=http://localhost:8080
```

### **Frontend Architecture**

```
Browser Request
    ↓
Next.js Middleware (proxy.ts)
    ↓ (Auth Check)
Route Handler
    ↓
Page/Component
    ↓ (useEffect)
API Client (lib/api.ts)
    ↓
Backend API
    ↓
Response ← Display in Component with Theme
```

---

## 📄 Root Level Files

### **Documentation Files**

#### **EKSEKUSI_PROYEK.md** (Indonesian: Project Execution)
- **Content**: Step-by-step setup and execution guide
- **Audience**: Developers setting up the project
- **Includes**: Environment setup, running backend/frontend, troubleshooting

#### **PROMPT_PANDUAN_PROYEK_WEB.md** (Architecture Guide)
- **Content**: Architecture decisions, design patterns, system prompts
- **Includes**: Project structure rationale, AI agent configuration
- **Use Case**: Understanding design decisions and extending project

#### **STEP7_LENGKAP.md** (Complete Code Templates)
- **Content**: Ready-to-use code snippets and templates
- **Includes**: Component examples, API patterns, common implementations
- **Use Case**: Quick reference for coding patterns

### **Utility Scripts**

#### **scripts/reset-dev.ps1** (PowerShell)
- **Purpose**: Reset development environment to clean state
- **Actions**:
  - Delete node_modules and package-lock.json
  - Delete Go cache files
  - Delete uploads directory
  - Reinitialize database
  - Reinstall dependencies
- **Usage**: `pwsh scripts/reset-dev.ps1` (from project root)

### **Other Root Files**

#### **.gitignore**
- **Excludes**: 
  - `node_modules/`, `dist/`, `build/`
  - `.env`, `.env.local`
  - Uploads, database files
  - IDE files (`.vscode/`, `.idea/`)
  - OS files (`Thumbs.db`, `.DS_Store`)

#### **uploads/** Directory
- **Runtime Created**: Stores user-uploaded files
- **Path**: `uploads/` (frontend and backend)
- **Security**: Files served through API with validation

---

## 💾 Data Models

### **Video Model**
```
ID:           uint (auto-incremented primary key)
CreatedAt:    timestamp
UpdatedAt:    timestamp
DeletedAt:    timestamp (soft delete)
Title:        string (required)
Subtitle:     string (optional)
Description:  string (optional, long text)
ThumbnailURL: string (image URL)
VideoURL:     string (video file URL)
PreviewURL:   string (preview/thumbnail URL)
Duration:     int64 (seconds)
Views:        int64 (view count counter)
IsPublished:  boolean (visibility flag)
```

**Relationships**: None defined yet (extensible for categories, tags)  
**Soft Deletes**: Uses `DeletedAt` field for logical deletion

### **Banner Model**
```
ID:        uint (auto-incremented primary key)
CreatedAt: timestamp
UpdatedAt: timestamp
DeletedAt: timestamp (soft delete)
Title:     string (banner title)
ImageURL:  string (promotional image URL)
LinkURL:   string (click destination)
IsActive:  boolean (visibility flag)
SortOrder: int (display order in slider)
```

**Use Cases**: Homepage promotional content, advertisements, announcements  
**Ordering**: Ordered by `SortOrder` field for predictable slider sequence

### **Ad Model**
```
ID:         uint (auto-incremented primary key)
CreatedAt:  timestamp
UpdatedAt:  timestamp
DeletedAt:  timestamp (soft delete)
Name:       string (internal ad name)
AdCode:     string (HTML content - ad creative)
DeviceType: string (mobile|desktop|all)
IsActive:   boolean (visibility flag)
```

**Device Targeting**: Serve different ads based on device type  
**HTML Support**: Allows inline scripts and rich media  
**Validation**: Should sanitize AdCode on frontend display

---

## 🔌 API Endpoints Reference

### **Public Endpoints** (No Authentication Required)

#### Videos
| Method | Endpoint | Query Params | Response |
|--------|----------|--------------|----------|
| GET | `/api/videos` | `limit`, `offset`, `search` | Array of videos |
| GET | `/api/videos/:id` | - | Single video object |
| GET | `/api/videos/featured` | - | Array of featured videos |

#### Content
| Method | Endpoint | Response |
|--------|----------|----------|
| GET | `/api/banners` | Array of active banners |
| GET | `/api/ads/serve` | Array of active ads |

### **Protected Endpoints** (Admin Authentication Required)

#### Authentication
| Method | Endpoint | Body | Response |
|--------|----------|------|----------|
| POST | `/api/admin/login` | `{username, password}` | `{token, message}` |
| POST | `/api/admin/logout` | - | Success message |

#### Video Management
| Method | Endpoint | Body | Response |
|--------|----------|------|----------|
| GET | `/api/admin/videos` | - | Array of all videos |
| POST | `/api/admin/videos` | Video object | Created video |
| PUT | `/api/admin/videos/:id` | Updated fields | Updated video |
| DELETE | `/api/admin/videos/:id` | - | Success message |

#### Banner Management
| Method | Endpoint | Body | Response |
|--------|----------|------|----------|
| GET | `/api/admin/banners` | - | Array of all banners |
| POST | `/api/admin/banners` | Banner object | Created banner |
| PUT | `/api/admin/banners/:id` | Updated fields | Updated banner |
| DELETE | `/api/admin/banners/:id` | - | Success message |

#### Ad Management
| Method | Endpoint | Body | Response |
|--------|----------|------|----------|
| GET | `/api/admin/ads` | - | Array of all ads |
| POST | `/api/admin/ads` | Ad object | Created ad |
| PUT | `/api/admin/ads/:id` | Updated fields | Updated ad |
| DELETE | `/api/admin/ads/:id` | - | Success message |

#### File Upload
| Method | Endpoint | Body | Response |
|--------|----------|------|----------|
| POST | `/api/admin/upload` | FormData (file) | `{url, filename}` |

---

## ⚙️ Environment Configuration

### **Backend Environment Variables** (.env)

```env
# Server Configuration
PORT=8080

# Database
DB_PATH=./videoplatform.db
UPLOAD_PATH=./uploads

# Authentication
JWT_SECRET=your_secret_key_here_minimum_32_chars
ADMIN_USERNAME=admin
ADMIN_PASSWORD=admin_password

# Security
COOKIE_SECURE=false    # Set to true for HTTPS
CORS_ORIGIN=http://localhost:3000

# Logging (optional)
LOG_LEVEL=info
```

**Required for Production**:
- `JWT_SECRET`: Generate a strong secret (32+ characters)
- `COOKIE_SECURE`: Set to `true` when using HTTPS
- `ADMIN_USERNAME` & `ADMIN_PASSWORD`: Change defaults
- `CORS_ORIGIN`: Set to actual frontend URL

### **Frontend Environment Variables** (.env.local)

```env
NEXT_PUBLIC_API_URL=http://localhost:8080
```

**Notes**:
- `NEXT_PUBLIC_*` prefix makes it available to browser
- Must point to backend API URL
- In production: `https://your-api-domain.com`

---

## 🚀 Development Workflow

### **Initial Setup**

1. **Clone and Navigate**
   ```bash
   cd d:\pgarcv
   ```

2. **Backend Setup**
   ```bash
   cd backend
   go mod tidy
   # Create .env file with configuration
   cp .env.example .env
   ```

3. **Frontend Setup**
   ```bash
   cd frontend
   npm install
   # Create .env.local with API URL
   ```

### **Running Development Servers**

**Terminal 1 - Backend (Go)**
```bash
cd backend
go run ./cmd/server/main.go
# Server runs on http://localhost:8080
```

**Terminal 2 - Frontend (Next.js)**
```bash
cd frontend
npm run dev
# App runs on http://localhost:3000
```

### **Building for Production**

**Backend**
```bash
cd backend
go build -o server ./cmd/server
./server
```

**Frontend**
```bash
cd frontend
npm run build
npm run start
```

### **Development Tools**

#### **Resetting Development Environment**
```bash
# PowerShell
pwsh scripts/reset-dev.ps1
```

This script:
- Clears Node modules and dependencies
- Resets database to initial state
- Clears upload directory
- Reinstalls all dependencies

### **Debugging**

**Backend (Go)**
- Check logs in terminal
- Use `fmt.Println()` for debugging
- Verify `.env` file configuration
- Check SQLite database: `sqlite3 videoplatform.db`

**Frontend (Next.js)**
- Use browser DevTools (F12)
- Check terminal for build errors
- Verify `.env.local` configuration
- Use `console.log()` in components

### **Database Management**

**Access SQLite Database**
```bash
cd backend
sqlite3 videoplatform.db
.tables              # List all tables
SELECT * FROM videos;  # Query data
.quit                # Exit
```

**Reset Database**
- Delete `videoplatform.db`
- Restart backend server (auto-creates)
- Data re-seeded from `database.go`

---

## 📊 Project Statistics

### **File Counts**
- **Backend**: ~12 Go files
- **Frontend**: ~25 TypeScript/TSX files
- **Configuration**: ~8 config files
- **Total**: ~50+ files

### **Lines of Code** (Estimated)
- **Backend**: ~2,000+ lines
- **Frontend**: ~3,000+ lines
- **Total**: ~5,000+ lines

### **Database**
- **Tables**: 3 (videos, banners, ads)
- **Fields**: ~35 total
- **Relationships**: Flat structure (no foreign keys yet)

### **Routes**
- **Public Routes**: 6 endpoints
- **Protected Routes**: 12 endpoints
- **Page Routes**: 10 pages (public + admin)
- **Component Routes**: Dynamic routes with parameters

### **Completion Status**
- ✅ Core Features: 70% Complete
  - Video CRUD, viewing, search
  - Banner management
  - Ad serving
  - Admin authentication
  - Responsive design

- 🔄 Medium Priority: 20% Complete
  - Categories/Tags
  - Comments system
  - Advanced analytics
  - Video processing

- 📋 Future Features: 10% Planned
  - Subscription/Paywall
  - User accounts
  - Social features
  - API rate limiting

---

## 🔗 References & Quick Links

### **Documentation Files**
- [Backend Setup Guide](EKSEKUSI_PROYEK.md)
- [Architecture & Design](PROMPT_PANDUAN_PROYEK_WEB.md)
- [Code Templates](STEP7_LENGKAP.md)

### **Key Configuration Files**
- [Backend Config](backend/config/config.go)
- [Frontend Config](frontend/next.config.ts)
- [Package Dependencies](frontend/package.json)

### **Important Directories**
- Backend: [cmd/server/main.go](backend/cmd/server/main.go) - Entry point
- Frontend: [app/layout.tsx](frontend/app/layout.tsx) - Root layout
- Components: [components/](frontend/components/) - Reusable parts
- API: [lib/api.ts](frontend/lib/api.ts) - API client

---

## 📝 Notes

- **Database**: SQLite auto-migrates on backend startup
- **Uploads**: Stored in `uploads/` directory (both locations)
- **Authentication**: JWT stored in secure HTTP-only cookies
- **Styling**: CSS variables enable theme switching
- **CORS**: Configured for localhost:3000 by default
- **File Extensions**: `.env` files not committed to git

---

**Document Version**: 1.0  
**Last Updated**: May 10, 2026  
**Scope**: Complete project audit and structure documentation
