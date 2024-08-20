# FoodBuddy API

FoodBuddy is a restaurant aggregator platform built using Go, Gin, MySQL, Jenkins, AWS, and Cloudinary for image uploads. The platform allows users to search for restaurants, manage their profiles, order food, and more. The API is structured with different routes for users, restaurants, and administrators, ensuring a seamless experience for all parties involved.

## Key Features

### User Management
- **Profile Management:** Users can create and update their profiles, including personal details and preferences.
- **Favorites & Addresses:** Ability to save favorite dishes and delivery addresses for quick ordering.
- **Wallet Information:** Secure handling of user wallet information for seamless transactions.

### Restaurant Management
- **Menu Items:** Restaurants can easily add, update, or remove items from their menus.
- **Order Handling:** Streamlined process for receiving and updating the status of customer orders.
- **Promotions:** Capability to create and manage promotional offers to attract customers.

### Order Processing
- **Order Placement:** Users can browse menus and place orders with ease.
- **Payment Integration:** Secure payment processing through Stripe and Razorpay.
- **Order Tracking:** Real-time tracking of order status from placement to delivery.

### Administrative Control
- **User & Restaurant Management:** Admins can oversee user accounts and restaurant listings.
- **Category Management:** Organize restaurants and menu items into relevant categories.

### Authentication
- **Secure Access:** Robust authentication mechanisms for users, restaurants, and admins.

### Referral System
- **Engagement & Rewards:** A referral system to encourage user engagement and offer rewards.

## Technology Stack

- **Backend:** Powered by the Go programming language using the Gin Gonic framework.
- **Database:** Compatible with MySQL and other relational databases.
- **API:** RESTful API endpoints for seamless integration and interaction across the platform.

## Payment Integration

- **Stripe & Razorpay:** Integrated support for popular payment gateways for secure transactions.

## Installation

To set up the project locally, follow these steps:

1. **Clone the Repository:**

    ```bash
    git clone https://github.com/liju-github/FoodBuddy-API.git
    cd FoodBuddy-API
    ```

2. **Set Up the Environment Variables:**

    Create a `.env` file in the root directory and add the following variables:

    ```bash
    SERVERIP=localhost:8080
    CLIENTID=your_google_oauth_client_id
    CLIENTSECRET=your_google_oauth_client_secret
    DBUSER=your_database_username
    DBPASSWORD=your_database_password
    DBNAME=your_database_name
    JWTSECRET=your_jwt_secret_key
    CLOUDNAME=your_cloudinary_cloud_name
    CLOUDINARYACCESSKEY=your_cloudinary_access_key
    CLOUDINARYSECRETKEY=your_cloudinary_secret_key
    CLOUDINARYURL=your_cloudinary_url
    RAZORPAY_KEY_ID=your_razorpay_key_id
    RAZORPAY_KEY_SECRET=your_razorpay_key_secret
    SMTPAPP=your_smtp_app_password
    STRIPE_KEY=your_stripe_secret_key
    STRIPE_WEBHOOK_SECRET=your_stripe_webhook_secret
    ```

3. **Install Dependencies:**

    ```bash
    go mod tidy
    ```

4. **Run the Application:**

    ```bash
    go run .
    ```
## API Documentation

Detailed API documentation is available [here](https://documenter.getpostman.com/view/32055383/2sA3e488Sh).

---
