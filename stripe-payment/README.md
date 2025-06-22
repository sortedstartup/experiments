## Stripe Payment Integration (One-Time + Subscription)

This project supports both one-time payments and recurring subscriptions using Stripe Checkout. Below are the setup steps to get the payment feature running successfully.

---

### Prerequisites

- React.js and npm installed
- Go installed
- Stripe account ([https://stripe.com](https://stripe.com))
- Your Stripe API Keys

---

### Project Structure (High Level)
├── backend
├── api/ # Contains Stripe integration logic
├── db/ # SQLite DB logic (saving paid users)
├── main.go # Gin backend entry point
└──.env # Your backend secrets

├── frontend/ # React frontend (Vite based)
├── pages/ # ui setup
└── .env # Your frontend secrets

---

### Backend Setup

1. **Go to the backend folder:**
    ```bash
    cd backend
    ```

2. **Install dependencies**  
   ```bash
   go mod tidy
   ```

3. **Create .env file in the backend folder with the following variables:**
    ```env
    STRIPE_SECRET_KEY=sk_test_...
    STRIPE_WEBHOOK_SECRET=whsec_...
    STRIPE_SUBSCRIPTION_PRICE_ID=price_...
    FRONTEND_URL=http://localhost:5173
    ```

4. **Run the backend**
    ```bash
    go run main.go
    ```

---

### Frontend Setup

1. **Go to the frontend folder:**
    ```bash
    cd frontend
    ```
    
2. **Create a .env file in frontend folder:**
    ```bash
    VITE_BACKEND_URL=http://localhost:8080
    ```

3. **Install and run:**
    ```bash
    npm install
    npm run dev
    ```

---

### Test the Payment Flow

1. **Visit `http://localhost:5173`**

2. **Click on:**

   - "Buy Now" for one-time payment

   - "Subscribe Monthly" for subscription

3. **Complete checkout with ([Stripe test cards](https://docs.stripe.com/testing))**

---

### How It Works (Under the Hood)

- **Home.jsx triggers backend /checkout-session or /subscription-session API.**

- **Backend creates a Stripe session and returns sessionId.**

- **Frontend redirects to Stripe Checkout**

- **After payment, Stripe sends checkout.session.completed event to /webhook.**

- **Webhook parses session and saves user to SQLite (if paid).**

- **A new page /paid exists that is only accessible to verified paid users.**

---
### Need Help?

- **Stripe Docs: ([https://docs.stripe.com/](https://docs.stripe.com/))**

- **Test Cards: ([https://docs.stripe.com/testing](https://docs.stripe.com/testing))**

