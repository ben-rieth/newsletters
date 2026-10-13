// Held in memory only, never in storage that outlives the tab. A reload loses it,
// and the user signs in again, which also sends them a fresh code.
let pendingPassword: string | null = null;

export const setPendingVerificationPassword = (password: string) => {
  pendingPassword = password;
};

export const getPendingVerificationPassword = () => pendingPassword;

export const clearPendingVerificationPassword = () => {
  pendingPassword = null;
};
