# Hướng Dẫn Kiểm Thử Hệ Thống AI Clipper (End-to-End)

Tài liệu này hướng dẫn cách chạy và kiểm thử luồng xử lý video bất đồng bộ (Asynchronous Video Processing) sử dụng Go, RabbitMQ, Modal AI và Server-Sent Events (SSE).

## 1. Chuẩn Bị Môi Trường

### A. RabbitMQ
Hệ thống yêu cầu RabbitMQ để truyền tin nhắn giữa Server và Worker.
Đảm bảo bạn đã chạy RabbitMQ (thông qua Docker):

```powershell
# Tại thư mục gốc của dự án
docker-compose up -d rabbitmq
```
*Thông tin mặc định:* `amqp://admin:admin123@localhost:5672/`

### B. Biến Môi Trường
Đảm bảo file `.env` (hoặc biến môi trường hệ thống) trong thư mục `server2` đã được cấu hình đầy đủ:
- `SUPABASE_URL`, `SUPABASE_SERVICE_ROLE_KEY`
- `MODAL_URL`, `MODAL_TOKEN`
- `RABBITMQ_URL=amqp://admin:admin123@localhost:5672/`
- `JWT_SECRET`

---

## 2. Khởi Động Dịch Vụ

Bạn cần mở **3 cửa sổ Terminal** riêng biệt:

### Terminal 1: Chạy API Server
Server này chịu trách nhiệm nhận File Upload và đẩy SSE cho Client.
```powershell
cd server2
go run cmd/server/main.go
```
*Dấu hiệu thành công:* Log hiện `Server starting on :8080` và kết nối RabbitMQ thành công.

### Terminal 2: Chạy Worker
Worker này chịu trách nhiệm xử lý video ngầm.
```powershell
cd server2
go run cmd/worker/main.go
```
*Dấu hiệu thành công:* Log hiện `RabbitMQ Consumer created` và `Started consuming messages`.

### Terminal 3: Backend Python (Modal)
Đảm bảo bạn đã deploy code Python lên Modal:
```powershell
modal deploy BE/main.py
```
*Lấy URL kết quả deploy dán vào `MODAL_URL` trong file cấu hình của Go.*

---

## 3. Quy Trình Test (Các bước thực hiện)

Chúng ta sẽ giả lập Client bằng **Postman** (hoặc cURL).

### Bước 1: Đăng nhập (Lấy Token)
Gọi API Login để lấy `access_token`.
*   **Method:** `POST`
*   **URL:** `http://localhost:8080/auth/login`
*   **Body (JSON):**
    ```json
    {
      "email": "test@example.com",
      "password": "yourpassword"
    }
    ```
*   **Copy:** Lưu lại `access_token` và `user_id` trả về.

### Bước 2: Kết nối SSE (Lắng nghe trạng thái)
Mở một tab mới trong Postman (hoặc dùng trình duyệt/cURL) để lắng nghe sự kiện.
*   **Method:** `GET`
*   **URL:** `http://localhost:8080/api/v1/events`
*   **Headers:**
    *   `Authorization`: `Bearer <access_token>`
*   **Hành động:** Gửi request. Bạn sẽ thấy kết nối **không đóng lại** (Loading...) mà giữ trạng thái kết nối. Đây là đúng.

### Bước 3: Upload Video
Quay lại tab khác để upload file.
*   **Method:** `POST`
*   **URL:** `http://localhost:8080/api/v1/upload`
*   **Headers:**
    *   `Authorization`: `Bearer <access_token>`
*   **Body (form-data):**
    *   `file`: (Chọn file video .mp4 từ máy tính)
    *   `config`: `{"prompt": "funny moments", "clip_count": 3, "subtitle": true}` (Dạng Text/JSON)

*   **Kết quả mong đợi (Response):**
    ```json
    {
        "status": "queued",
        "message": "File uploaded and processing started",
        ...
    }
    ```

---

## 4. Xác Minh Log (Quan sát Real-time)

Ngay sau khi bấm Upload, hãy quan sát các Terminal:

1.  **Terminal 1 (Server):**
    *   Log: `File uploaded successfully`
    *   Log: `Video processing message published`
    *   (Sau một lúc) Log: `Received status update for user...`
    *   Log: `Sending SSE to user...`

2.  **Terminal 2 (Worker):**
    *   Log: `Received message from video_processing`
    *   Log: `Processing video for file...`
    *   Log: `Calling Modal for file...` (Lúc này Worker sẽ đợi Modal xử lý)
    *   (Sau khi Modal xong) Log: `Video processing completed successfully. Saved X clips.`
    *   Log: `Message processed successfully`

3.  **Terminal Client (Postman SSE):**
    *   Bạn sẽ thấy một event mới xuất hiện:
    ```json
    event: video_status
    data: {"file_id":"...", "status":"success", "clip_count": 5}
    ```

---

## 5. Troubleshooting (Xử lý lỗi thường gặp)

| Vấn đề | Kiểm tra | Giải pháp |
| :--- | :--- | :--- |
| **Server báo lỗi connect RabbitMQ** | Docker | Chạy `docker ps` xem rabbitmq có đang chạy không. Kiểm tra port 5672. |
| **Worker không nhận được tin nhắn** | Queue Name | Đảm bảo Server và Worker dùng chung `RABBITMQ_URL`. |
| **Modal trả về lỗi** | Python Log | Vào Dashboard Modal.com xem log chi tiết của app `ai-podcast-clipper`. |
| **0 Clips saved** | Logic | Có thể AI không tìm thấy khoảnh khắc phù hợp với prompt. Thử prompt khác phổ biến hơn. |
| **SSE không nhận được tin** | UserID | Đảm bảo Token dùng để kết nối SSE và Token dùng để Upload là **cùng một User**. |

---

## 6. Sơ đồ Luồng Dữ Liệu

1.  **Client** --(Upload)--> **Server** --(Publish)--> **RabbitMQ (Video Queue)**
2.  **Server** --(Response "Queued")--> **Client**
3.  **RabbitMQ** --(Consume)--> **Worker**
4.  **Worker** --(Call)--> **Modal AI** --(Return Clips)--> **Worker**
5.  **Worker** --(Save DB)--> **PostgreSQL**
6.  **Worker** --(Publish)--> **RabbitMQ (Status Queue)**
7.  **RabbitMQ** --(Consume)--> **Server**
8.  **Server** --(SSE Push)--> **Client**
