# PetPlate

PetPlate is an online platform designed to cater to pet owners' needs by offering a wide range of pet food products and services. Built using Go, the Gin framework, and PostgreSQL, PetPlate provides a seamless experience for users and admins, with robust features such as authentication, product management, order and booking systems, and much more. The platform ensures a user-friendly experience, secure transactions, and efficient service management.

## Key Features
   - **User Features:**
        Product Browsing: Explore pet food and supplies by categories.
        Wishlist and Cart: Add products to a wishlist or cart for easy ordering.
        Secure Order Management: Place and track orders with real-time status updates.
        Service Booking: Book pet-related services like grooming and veterinary care with flexible scheduling.
        Profile Management: Update personal details and view order history.
        Authentication: Secure user signup/login with:
        OTP verification.
        Google authentication.
        JWT-based session management.
   - **Admin Features:**
    User Management: View, block, or unblock users to maintain platform integrity.
    Product Management: Add, edit, or delete products with detailed categories.
    Order and Service Management: Oversee product orders and service bookings, update statuses, or cancel requests.
    Reports: Generate and export sales and booking reports for analysis.
    Payment Management: Manage payment methods and track successful transactions.

   - **Additional Highlights**

    Environment Configuration: Securely manage sensitive credentials using .env files.
    API Design: Modular and scalable routes for both users and admins.
    Hosting Ready: Deployable on cloud platforms like AWS or Render.
## Installation
    To set up the project locally, follow these steps:
- **1.Clone the Repository:**
        ```bash
        git clone https://github.com/your-username/PetPlate.git
        cd PetPlate
        ```
- **2.Set Up the Environment Variables:**
    ```bash
     Create a .env file in the root directory and add the following variables:
        DB_HOST=127.0.0.1
        DB_USER=your_database_username
        DB_PASSWORD=your_database_password
        DB_NAME=your_database_name
        DB_PORT=5432
        JWT_SECRET=your_jwt_secret_key
        GOOGLE_CLIENT_ID=your_google_client_id
        GOOGLE_CLIENT_SECRET=your_google_client_secret
        ```
- **3.Install Dependencies:**
    ```bash
            go mod tidy
    ```
- **4.Run the Application:**
    ```bash
        go run .
        ```
## API Documentation
    Comprehensive API documentation is available
    [here](https://docs.google.com/document/d/1QwAFAD_KwVfsD837PuTd94OSrIP7n9K-H-hRfyFvNYM/edit?tab=t.0#heading=h.z337ya)
