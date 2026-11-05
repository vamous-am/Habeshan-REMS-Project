export const ETHIOPIAN_PHONE_PATTERN = /^(?:0[79]\d{8}|\+251[79]\d{8})$/;

export const ETHIOPIAN_PHONE_MESSAGE =
  "Enter a valid Ethiopian phone number";

export function isValidEthiopianPhone(phone: string): boolean {
  return ETHIOPIAN_PHONE_PATTERN.test(phone.trim());
}
