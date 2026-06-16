export type ToastType = "info" | "success" | "warning" | "error" | "joined" | "left" | "track";

export interface Toast {
	id: string;
	message: string;
	type: ToastType;
	duration?: number;
}

let toastsList = $state<Toast[]>([]);

export const toasts = {
	get list() {
		return toastsList;
	},

	add(message: string, type: ToastType = "info", duration = 4000) {
		const id = Math.random().toString(36).substring(2, 9);
		const newToast: Toast = { id, message, type, duration };
		toastsList = [...toastsList, newToast];

		if (duration > 0) {
			setTimeout(() => {
				this.remove(id);
			}, duration);
		}
	},

	remove(id: string) {
		toastsList = toastsList.filter((t) => t.id !== id);
	},

	success(message: string, duration?: number) {
		this.add(message, "success", duration);
	},

	info(message: string, duration?: number) {
		this.add(message, "info", duration);
	},

	warning(message: string, duration?: number) {
		this.add(message, "warning", duration);
	},

	error(message: string, duration?: number) {
		this.add(message, "error", duration);
	}
};
