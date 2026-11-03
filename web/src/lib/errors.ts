import type { components } from '#/api/schema';

type ErrorDetail = components['schemas']['ErrorDetail'];

const REQUEST_PARTS = new Set(['body', 'path', 'query', 'header', 'cookie']);

const capitalize = (text: string) =>
  text.charAt(0).toUpperCase() + text.slice(1);

const toFieldLabel = (field: string) =>
  capitalize(field.replace(/([a-z])([A-Z])/g, '$1 $2').toLowerCase());

const formatErrorDetail = ({ location, message = '' }: ErrorDetail) => {
  const field = location?.split('.').at(-1);
  if (!field || REQUEST_PARTS.has(field)) {
    return capitalize(message);
  }
  return `${toFieldLabel(field)} ${message}`;
};

const getValidationMessages = (error: object): string | undefined => {
  if (!('errors' in error) || !Array.isArray(error.errors)) {
    return undefined;
  }
  const messages = (error.errors as ErrorDetail[])
    .filter((detail) => detail.message)
    .map(formatErrorDetail);
  return messages.length > 0 ? messages.join('\n') : undefined;
};

export const getErrorMessage = (error: unknown): string => {
  if (error instanceof Error) {
    return error.message;
  }
  if (typeof error !== 'object' || error === null) {
    return 'Something went wrong';
  }
  const validationMessages = getValidationMessages(error);
  if (validationMessages) {
    return validationMessages;
  }
  if ('detail' in error && typeof error.detail === 'string') {
    return error.detail;
  }
  return 'Something went wrong';
};
