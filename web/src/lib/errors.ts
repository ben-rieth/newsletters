export const getErrorMessage = (error: unknown): string => {
  if (error instanceof Error) {
    return error.message;
  }
  if (
    typeof error === 'object' &&
    error !== null &&
    'detail' in error &&
    typeof error.detail === 'string'
  ) {
    return error.detail;
  }
  return 'Something went wrong';
};
