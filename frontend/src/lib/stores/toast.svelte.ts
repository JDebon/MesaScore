interface Toast {
	id: number;
	message: string;
	type: 'success' | 'error' | 'info';
}

let toasts = $state<Toast[]>([]);
let nextId = 0;

export function getToasts(): Toast[] {
	return toasts;
}

export function addToast(message: string, type: Toast['type'] = 'success') {
	const id = nextId++;
	toasts = [...toasts, { id, message, type }];
	setTimeout(() => dismissToast(id), 3000);
}

export function dismissToast(id: number) {
	toasts = toasts.filter((t) => t.id !== id);
}
