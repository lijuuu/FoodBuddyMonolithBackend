# 🍔 FoodBuddy - Monolith Backend

FoodBuddy is a monolithic backend service for a food ordering application. It handles user authentication (with Google OAuth), restaurant and menu management, orders, payment processing, image uploads, and email notifications. It integrates major services like Stripe, Razorpay, Cloudinary, and SMTP to provide a complete backend solution for food delivery platforms.

---

## 🧩 Features

- Google OAuth-based user authentication  
- JWT-secured sessions  
- Restaurant, menu, and food item management  
- Add to cart, order placement, and order tracking  
- Cloudinary integration for image uploads  
- Stripe and Razorpay integration for secure payments  
- SMTP-based email sending (OTP, notifications, etc.)  
- Admin-level management of users, restaurants, and categories  

---

## 🛠 Technologies Used

* **Go** – Gin Web Framework
* **MySQL** – Relational DB with GORM ORM
* **Docker** – Containerized deployment
* **Google OAuth2** – Social login
* **Stripe & Razorpay** – Payment gateways
* **Cloudinary** – Media upload/storage
* **JWT** – Token-based authentication
* **SMTP (Gmail App)** – Email notifications
* **Kubernetes (Optional)** – Deployment (k8s manifests included)

---

## ⚙️ Requirements

Make sure you have the following installed:

### System Requirements

- **Go**: v1.20+  
- **Docker**: v20+  
- **MySQL**: v8+  
- **Git**: v2.30+  
- **SMTP-compatible email account** (e.g., Gmail App Password)  
- **Stripe + Razorpay accounts**  
- **Google Developer credentials** for OAuth  
- **Cloudinary account**

### Go Dependencies

To download all Go modules:

```bash
go mod tidy
````

---

## 🐳 Docker Setup

Clone the repository and build/run the Docker container.

### 1. Clone the project

```bash
git clone https://github.com/lijuuu/FoodBuddyMonolithBackend.git
cd FoodBuddyMonolithBackend
```

### 2. Create `.env` file

Use the `.env.example` file provided below to create your own `.env`.

```bash
cp .env.example .env
```

Edit `.env` and update the values with your own credentials.

### 3. Build Docker image

```bash
docker build -t foodbuddy-backend .
```

### 4. Run container

```bash
docker run -d --name foodbuddy \
  --env-file .env \
  -p 8080:8080 \
  foodbuddy-backend
```

> App should now be available at: `http://localhost:8080`

---

## 📦 Environment Variables

| Key                     | Description                                     |
| ----------------------- | ----------------------------------------------- |
| `SERVERIP`              | App base URL with port (e.g., `localhost:8080`) |
| `PORT`                  | Port to run the backend on                      |
| `CLIENTID`              | Google OAuth Client ID                          |
| `CLIENTSECRET`          | Google OAuth Client Secret                      |
| `DBUSER`                | MySQL database username                         |
| `DBPASSWORD`            | MySQL database password                         |
| `DBNAME`                | MySQL database name                             |
| `JWTSECRET`             | Secret for signing JWT tokens                   |
| `CLOUDNAME`             | Cloudinary cloud name                           |
| `CLOUDINARYACCESSKEY`   | Cloudinary API key                              |
| `CLOUDINARYSECRETKEY`   | Cloudinary API secret                           |
| `CLOUDINARYURL`         | Cloudinary URL                                  |
| `RAZORPAY_KEY_ID`       | Razorpay key ID                                 |
| `RAZORPAY_KEY_SECRET`   | Razorpay key secret                             |
| `SMTPAPP`               | App password for SMTP email (Gmail, etc.)       |
| `STRIPE_KEY`            | Stripe secret key                               |
| `STRIPE_WEBHOOK_SECRET` | Stripe webhook secret                           |

---

## 🧪 .env.example

```env
SERVERIP=localhost:8080
PORT=8080

CLIENTID=your_google_oauth_client_id
CLIENTSECRET=your_google_oauth_client_secret

DBUSER=root
DBPASSWORD=your_db_password
DBNAME=foodbuddy

JWTSECRET=your_jwt_secret

CLOUDNAME=your_cloudinary_cloud_name
CLOUDINARYACCESSKEY=your_cloudinary_api_key
CLOUDINARYSECRETKEY=your_cloudinary_api_secret
CLOUDINARYURL=cloudinary://your_cloudinary_key:your_cloudinary_secret@your_cloudinary_name

RAZORPAY_KEY_ID=your_razorpay_key_id
RAZORPAY_KEY_SECRET=your_razorpay_key_secret

SMTPAPP=your_smtp_app_password

STRIPE_KEY=your_stripe_secret_key
STRIPE_WEBHOOK_SECRET=your_stripe_webhook_secret
```

---

## 🧪 API Reference

You can find the full API documentation here:
👉 [Postman Collection (View Only)](https://documenter.getpostman.com/view/32055383/2sA3e488Sh)

Use this to explore all routes, methods, request/response formats.


## 👨‍💻 Author

**Liju Thomas**
[GitHub](https://github.com/lijuuu)

---
