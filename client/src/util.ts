export function debounce(fn: (...args: any[]) => void, delay: number) {
  let timeout: number;
  return (...args: any[]) => {
    clearTimeout(timeout);
    timeout = setTimeout(() => fn(...args), delay);
  };
}

export function onlineBinarySearch(arrLength: number, target: number, getValue: (val: number) => number, compare: (currVal: number, target: number) => number) {
  let lower = 0;
  let upper = arrLength - 1;
  while (lower < upper) {
    const mid = Math.floor((lower + upper) / 2);
    const currVal = getValue(mid);
    const comp = compare(currVal, target);
    if (comp < 0) {
      lower = mid + 1;
    } else if (comp > 0) {
      upper = mid - 1;
    } else {
      return mid;
    }
  }

  return lower;
}