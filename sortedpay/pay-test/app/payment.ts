import { atom } from "nanostores";
import {
     PaymentServiceClient, ListProductsRequest, Product, CreateStripeCheckoutSessionRequest, CreateRazorpayCheckoutSessionRequest, Currency, PaymentType, Interval, CreateStripeSubscriptionCheckoutSessionRequest, CreateRazorpaySubscriptionCheckoutSessionRequest, CheckUserProductAccessRequest
} from "../proto/paymentservice"
// import { createAuthenticatedClientOptions } from "../../lib/auth";
import { toast } from "sonner";

const client = new PaymentServiceClient("http://localhost:8080", {}, {
    format: "text"
});




export const ProductList = atom<Product[]>([]);

export const listProducts = async () => {
    try {
        const req = new ListProductsRequest({});
        const res = await client.ListProducts(req, {});
        ProductList.set(res.products);
        return res.products;
    } catch (err) {
        toast.error("Failed to list products");
        throw err;
    }
}

export const createStripeCheckoutSession = async (productId: string) => {
    try {
        const req = new CreateStripeCheckoutSessionRequest({ product_id: productId });
        const res = await client.CreateStripeCheckoutSession(req, {});
        toast.success("Checkout session created successfully");
        return res.session_url;
    } catch (err) {
        toast.error("Failed to create checkout session");
        throw err;
    }
}

export const createRazorpayCheckoutSession = async (productId: string) => {
    try {
        const req = new CreateRazorpayCheckoutSessionRequest({ product_id: productId });
        const res = await client.CreateRazorpayCheckoutSession(req, {});
        return {
            orderId: res.order_id,
            amount: res.amount,
            currency: res.currency
        };
    } catch (err) {
        toast.error("Failed to create checkout session");
        throw err;
    }
}

// New subscription methods
export const createStripeSubscriptionCheckoutSession = async (productId: string) => {
    try {
        const req = new CreateStripeSubscriptionCheckoutSessionRequest({
            product_id: productId
        });
        const res = await client.CreateStripeSubscriptionCheckoutSession(req, {});
        toast.success("Subscription checkout session created successfully");
        return res.session_url;
    } catch (err) {
        toast.error("Failed to create subscription checkout session");
        throw err;
    }
}

export const createRazorpaySubscriptionCheckoutSession = async (productId: string) => {
    try {
        const req = new CreateRazorpaySubscriptionCheckoutSessionRequest({
            product_id: productId
        });
        const res = await client.CreateRazorpaySubscriptionCheckoutSession(req, {});
        toast.success("Subscription checkout session created successfully");
        return {
            subscriptionId: res.subscription_id,
            amount: res.amount,
            currency: res.currency
        };

    } catch (err) {
        toast.error("Failed to create subscription checkout session");
        throw err;
    }
}

export const checkUserProductAccess = async (productId: string) => {
    try {
        const req = new CheckUserProductAccessRequest({ product_id: productId });
        const res = await client.CheckUserProductAccess(req, {});
        return res.has_access;
    } catch (err) {
        toast.error("Failed to check user product access");
        throw err;
    }
}