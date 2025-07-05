export function success(message: string) {
  console.log(`✅ ${message}`);
}

export function error(error: unknown) {
  console.log(`❌ ${error}`);
}

export function info(message: string) {
  console.log(`👍 ${message}`);
}
