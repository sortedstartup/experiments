import React, { useEffect, useState } from "react";
import { useStore } from "@nanostores/react";
import { useParams, useNavigate } from "react-router-dom";
import { TransactionsList, getTransactions } from "./store/payment";

const Transactions: React.FC = () => {
    const transactions = useStore(TransactionsList);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState("");
    const { page } = useParams<{ page: string }>();
    const navigate = useNavigate();
    const currentPage = parseInt(page || "1", 10);
    const pageSize = 10;

    const fetchTransactions = async (page: number) => {
        try {
            setLoading(true);
            setError("");
            const fetchedTransactions = await getTransactions(page, pageSize);
            TransactionsList.set(fetchedTransactions);
        } catch (err: any) {
            setError(err?.message || "Failed to fetch transactions");
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        // Redirect to page 1 if no page or invalid page
        if (!page || isNaN(currentPage) || currentPage < 1) {
            navigate("/transactions/1", { replace: true });
            return;
        }
        fetchTransactions(currentPage);
    }, [page, currentPage, navigate]);

    const handlePreviousPage = () => {
        if (currentPage > 1) {
            navigate(`/transactions/${currentPage - 1}`);
        }
    };

    const handleNextPage = () => {
        if (transactions.length === pageSize) {
            navigate(`/transactions/${currentPage + 1}`);
        }
    };

    const formatAmount = (amount: number, currency: string) => {
        return new Intl.NumberFormat(undefined, {
            style: "currency",
            currency: currency || "USD",
        }).format(amount / 100);
    };

    const formatDate = (dateString: string) => {
        return new Date(dateString).toLocaleDateString(undefined, {
            year: 'numeric',
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit'
        });
    };

    if (loading) {
        return (
            <div className="p-6">
                <h1 className="text-2xl font-semibold mb-6">Transactions</h1>
                <div className="text-center py-8">Loading transactions...</div>
            </div>
        );
    }

    if (error) {
        return (
            <div className="p-6">
                <h1 className="text-2xl font-semibold mb-6">Transactions</h1>
                <div className="text-red-600 text-center py-8">Error: {error}</div>
            </div>
        );
    }

    return (
        <div className="p-6">
            <h1 className="text-2xl font-semibold mb-6">Transactions</h1>
            
            {transactions.length === 0 ? (
                <div className="text-gray-500 text-center py-8">No transactions found</div>
            ) : (
                <>
                    {/* Transaction List */}
                    <div className="space-y-4">
                        {transactions.map((transaction) => (
                            <div key={transaction.id} className="border rounded-lg p-4 bg-white shadow-sm">
                                <div className="flex justify-between items-start">
                                    <div className="flex-1">
                                        <h3 className="font-medium text-lg">
                                            {transaction.productName || 'Unknown Product'}
                                        </h3>
                                        <p className="text-sm text-gray-500 mt-1">
                                            Transaction ID: {transaction.id}
                                        </p>
                                        <p className="text-sm text-gray-500">
                                            User ID: {transaction.userId}
                                        </p>
                                        <p className="text-sm text-gray-500">
                                            Date: {formatDate(transaction.createdAt)}
                                        </p>
                                    </div>
                                    <div className="text-right">
                                        <p className="font-semibold text-lg">
                                            {formatAmount(Number(transaction.amount), transaction.currency)}
                                        </p>
                                        <p className={`text-sm font-medium ${
                                            transaction.status.toLowerCase() === 'completed' || 
                                            transaction.status.toLowerCase() === 'success' ||
                                            transaction.status.toLowerCase() === 'paid'
                                                ? 'text-green-600' 
                                                : transaction.status.toLowerCase() === 'pending'
                                                ? 'text-yellow-600'
                                                : 'text-red-600'
                                        }`}>
                                            {transaction.status}
                                        </p>
                                    </div>
                                </div>
                            </div>
                        ))}
                    </div>
                    
                    {/* Pagination */}
                    <div className="flex justify-center items-center mt-8 space-x-2">
                        <button 
                            onClick={handlePreviousPage}
                            disabled={currentPage === 1}
                            className="px-3 py-2 border rounded hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                        >
                            &lt;
                        </button>
                        <span className="px-3 py-2 bg-blue-500 text-white rounded">
                            {currentPage}
                        </span>
                        <button 
                            onClick={handleNextPage}
                            disabled={transactions.length < pageSize}
                            className="px-3 py-2 border rounded hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                        >
                            &gt;
                        </button>
                    </div>
                </>
            )}
        </div>
    );
};

export default Transactions;
