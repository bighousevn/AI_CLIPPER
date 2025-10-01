# 🤖 AI Clipper

Dự án AI Clipper là một ứng dụng web cho phép người dùng tạo các video clip ngắn từ các video dài trên YouTube một cách tự động bằng công nghệ AI. Người dùng có thể dán link YouTube, và hệ thống sẽ tự động xử lý, phân tích và tạo ra các clip nổi bật.

## ✨ Tính Năng Chính

- **Đăng ký / Đăng nhập:** Hệ thống xác thực người dùng an toàn.
- **Dashboard người dùng:** Giao diện quản lý các video đã xử lý và các clip đã tạo.
- **Nhập video từ YouTube:** Dán URL video YouTube để hệ thống tải về và xử lý.
- **Tự động tạo clip (AI-Powered):** Lõi AI tự động phân tích và cắt các phân đoạn hấp dẫn từ video gốc.
- **Quản lý và tải clip:** Xem, quản lý và tải các clip đã được tạo ra.

## 🏗️ Kiến Trúc Hệ Thống

Dự án được xây dựng theo kiến trúc microservices, bao gồm các thành phần chính:

1.  **Frontend (Next.js):** Giao diện người dùng được xây dựng bằng Next.js và React, cung cấp trải nghiệm mượt mà và hiện đại. Giao tiếp với Backend Go qua các API.
2.  **Backend (Go):** Là API Gateway chính, xử lý các tác vụ liên quan đến người dùng như xác thực, quản lý thông tin người dùng, và giao tiếp với dịch vụ Python.
3.  **AI & Downloader Service (Python):** Một dịch vụ riêng biệt viết bằng Python, chịu trách nhiệm tải video từ YouTube và chạy các thuật toán AI để xử lý, phân tích và tạo clip.
4.  **Database (Supabase):** Sử dụng Supabase (PostgreSQL) để lưu trữ dữ liệu người dùng, thông tin video, và các metadata liên quan.

- [System Architecture](docs/system.drawio.png)


- [Database schema](https://dbdiagram.io/d/DB-DA1-680f23941ca52373f59993fd)

## 🛠️ Công Nghệ Sử Dụng

- **Frontend:**

  - [Next.js](https://nextjs.org/)
  - [TypeScript](https://www.typescriptlang.org/)
  - [Tailwind CSS](https://tailwindcss.com/)

- **Backend (API Gateway):**

  - [Go (Golang)](https://golang.org/)
  - [GORM](https://gorm.io/)

- **Backend (AI Service):**

  - [Python](https://www.python.org/)

- **Database:**
  - [Supabase](https://supabase.io/)
  - [PostgreSQL](https://www.postgresql.org/)

## 🚀 Hướng Dẫn Cài Đặt và Chạy Dự Án

### Điều kiện cần có

- [Node.js](https://nodejs.org/en/) (v18 trở lên)
- [Go](https://golang.org/doc/install/) (v1.20 trở lên)
- [Python](https://www.python.org/downloads/) (v3.9 trở lên)
- [Docker](https://www.docker.com/products/docker-desktop/) (Tùy chọn, cho database)

### Các bước cài đặt

1.  **Clone repository:**

    ```bash
    git clone <your-repository-url>
    cd AI_CLIPPER
    ```

2.  **Thiết lập biến môi trường:**

    - Sao chép các file `.env.example` thành `.env` trong các thư mục `FE`, `server`, và `BE`.
    - Điền các thông tin cần thiết như chuỗi kết nối database, API keys, etc.

3.  **Chạy Frontend (Thư mục `FE`):**

    ```bash
    cd FE
    npm install
    npm run dev
    ```

    Frontend sẽ chạy tại `http://localhost:3000`.

4.  **Chạy Backend Go (Thư mục `server`):**

    ```bash
    cd ../server
    go mod tidy
    go run main.go
    ```

    Backend Go sẽ chạy tại `http://localhost:8080` (hoặc cổng bạn cấu hình).

5.  **Chạy Backend Python (Thư mục `BE`):**
    ```bash
    cd ../BE
    pip install -r requirements.txt
    python main.py
    ```
    Backend Python sẽ chạy tại `http://localhost:5000` (hoặc cổng bạn cấu hình).

## 🤝 Đóng Góp

Chúng tôi hoan nghênh mọi sự đóng góp! Vui lòng tạo Pull Request hoặc mở Issue để thảo luận về các thay đổi bạn muốn thực hiện.

## 📄 Giấy Phép

Dự án này được cấp phép theo [MIT License](LICENSE).
