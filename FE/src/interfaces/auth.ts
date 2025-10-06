export interface LoginResponse {
    access_token: string;
    message: string;
}

export interface UserResponse {
    id: number;                 // ID người dùng (backend trả kiểu number)
    username: string;           // Tên đăng nhập
    email: string;              // Email người dùng
    credits: number;            // Số credits còn lại
    stripe_customer_id: string | null; // ID khách hàng Stripe (nếu có)
}
