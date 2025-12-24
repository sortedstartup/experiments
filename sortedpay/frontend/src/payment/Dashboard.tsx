
import { useEffect, useState } from "react";
import { useStore } from "@nanostores/react";
import { DashboardDataList, getDashboardData } from "./store/payment";

function Dashboard() {
    const dashboardDataList = useStore(DashboardDataList);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState("");

    useEffect(() => {
        const fetchDashboardData = async () => {
            try {
                setLoading(true);
                setError("");
                await getDashboardData();
            } catch (err: any) {
                setError(err?.message || "Failed to fetch dashboard data");
            } finally {
                setLoading(false);
            }
        };

        fetchDashboardData();
    }, []);

    const formatCurrency = (amount: number, currency: string) => {
        return new Intl.NumberFormat(undefined, {
            style: "currency",
            currency: currency,
        }).format(amount);
    };

    if (loading) {
        return (
            <div className="p-6">
                <h1 className="text-2xl font-semibold mb-6">Dashboard</h1>
                <div className="text-center py-8">Loading dashboard data...</div>
            </div>
        );
    }

    if (error) {
        return (
            <div className="p-6">
                <h1 className="text-2xl font-semibold mb-6">Dashboard</h1>
                <div className="text-red-600 text-center py-8">Error: {error}</div>
            </div>
        );
    }

    return (
        <div className="p-6">
            <h1 className="text-2xl font-semibold mb-6">Dashboard</h1>
            
            {dashboardDataList.length === 0 ? (
                <div className="text-gray-500 text-center py-8">No sales data available</div>
            ) : (
                <div className="space-y-6">
                    {dashboardDataList.map((currencyData) => (
                        <div key={currencyData.currency} className="bg-white rounded-lg shadow-sm border p-6">
                            <h2 className="text-lg font-semibold mb-4 text-gray-800">
                                {currencyData.currency} Sales
                            </h2>
                            
                            {/* Sales Row for this currency */}
                            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                                {/* Daily Sales */}
                                <div className="text-center">
                                    <p className="text-sm font-medium text-gray-600 mb-2">Today's Sales</p>
                                    <p className="text-2xl font-bold text-gray-900">
                                        {formatCurrency(currencyData.dailySales, currencyData.currency)}
                                    </p>
                                </div>

                                {/* Weekly Sales */}
                                <div className="text-center">
                                    <p className="text-sm font-medium text-gray-600 mb-2">Weekly Sales</p>
                                    <p className="text-2xl font-bold text-gray-900">
                                        {formatCurrency(currencyData.weeklySales, currencyData.currency)}
                                    </p>
                                </div>

                                {/* Monthly Sales */}
                                <div className="text-center">
                                    <p className="text-sm font-medium text-gray-600 mb-2">Monthly Sales</p>
                                    <p className="text-2xl font-bold text-gray-900">
                                        {formatCurrency(currencyData.monthlySales, currencyData.currency)}
                                    </p>
                                </div>
                            </div>
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
}

export default Dashboard;