import React from "react";

const Transactions: React.FC = () => {
    return (
        <div className="p-6">
            <h1 className="text-2xl font-semibold mb-6">Transactions</h1>
            
            {/* Transaction List */}
            <div className="space-y-4">
                {/* Sample Transaction Items */}
                <div className="border rounded-lg p-4 bg-white shadow-sm">
                    <div className="flex justify-between items-center">
                        <div>
                            <h3 className="font-medium">Product Purchase</h3>
                            <p className="text-sm text-gray-500">Transaction ID: #12345</p>
                        </div>
                        <div className="text-right">
                            <p className="font-semibold">$29.99</p>
                            <p className="text-sm text-green-600">Completed</p>
                        </div>
                    </div>
                </div>
                
                <div className="border rounded-lg p-4 bg-white shadow-sm">
                    <div className="flex justify-between items-center">
                        <div>
                            <h3 className="font-medium">Subscription Payment</h3>
                            <p className="text-sm text-gray-500">Transaction ID: #12346</p>
                        </div>
                        <div className="text-right">
                            <p className="font-semibold">$9.99</p>
                            <p className="text-sm text-green-600">Completed</p>
                        </div>
                    </div>
                </div>
            </div>
            
            {/* Pagination */}
            <div className="flex justify-center items-center mt-8 space-x-2">
                <button className="px-3 py-2 border rounded hover:bg-gray-50">
                    &lt;
                </button>
                <button className="px-3 py-2 bg-blue-500 text-white rounded">
                    1
                </button>
                <button className="px-3 py-2 border rounded hover:bg-gray-50">
                    &gt;
                </button>
            </div>
        </div>
    );
};

export default Transactions;
