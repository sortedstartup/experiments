import { useEffect, useRef } from "react";
import { createPortal } from "react-dom";

interface UserActionsMenuProps {
    userId: string;
    isOpen: boolean;
    onClose: () => void;
    onChangeRole: () => void;
    onRemove: () => void;
    buttonRef: React.RefObject<HTMLButtonElement>;
}

export default function UserActionsMenu({
    isOpen,
    onClose,
    onChangeRole,
    onRemove,
    buttonRef,
}: UserActionsMenuProps) {
    const menuRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        const handleClickOutside = (event: MouseEvent) => {
            if (menuRef.current && !menuRef.current.contains(event.target as Node) &&
                buttonRef.current && !buttonRef.current.contains(event.target as Node)) {
                onClose();
            }
        };

        if (isOpen) {
            document.addEventListener('mousedown', handleClickOutside);
        }

        return () => {
            document.removeEventListener('mousedown', handleClickOutside);
        };
    }, [isOpen, onClose, buttonRef]);

    if (!isOpen || !buttonRef.current) return null;

    const rect = buttonRef.current.getBoundingClientRect();

    return createPortal(
        <div
            ref={menuRef}
            className="fixed w-56 bg-gray-800 border border-gray-700 rounded-lg shadow-xl z-[9999]"
            style={{
                top: `${rect.bottom + 4}px`,
                left: `${rect.right - 224}px`, // 224px = 56 * 4 (w-56 in pixels)
            }}
        >
            <div className="py-1">
                <button
                    onClick={onChangeRole}
                    className="w-full flex items-center gap-3 px-4 py-2.5 text-sm text-gray-200 hover:bg-gray-700 transition-colors"
                >
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
                    </svg>
                    Change Role
                </button>
                <button
                    onClick={onRemove}
                    className="w-full flex items-center gap-3 px-4 py-2.5 text-sm text-red-400 hover:bg-gray-700 transition-colors"
                >
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7a4 4 0 11-8 0 4 4 0 018 0zM9 14a6 6 0 00-6 6v1h12v-1a6 6 0 00-6-6z" />
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M18 8l-4 4m0-4l4 4" />
                    </svg>
                    Remove from Tenant
                </button>
            </div>
        </div>,
        document.body
    );
}

