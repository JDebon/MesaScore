let pendingInviteCount = $state(0);

export function getPendingInviteCount(): number {
	return pendingInviteCount;
}

export function setPendingInviteCount(n: number) {
	pendingInviteCount = n;
}
