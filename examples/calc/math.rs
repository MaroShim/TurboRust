// Math module for Turbo Rust calculation suite

/// Computes n! (factorial) iteratively.
pub fn factorial(n: u64) -> u64 {
    if n <= 1 {
        return 1;
    }
    let mut res = 1;
    for i in 2..=n {
        res *= i;
    }
    res
}

/// Computes the greatest common divisor using Euclidean algorithm.
pub fn gcd(mut a: u64, mut b: u64) -> u64 {
    while b != 0 {
        let t = b;
        b = a % b;
        a = t;
    }
    a
}

/// Checks if a positive integer is prime.
pub fn is_prime(n: u64) -> bool {
    if n <= 1 {
        return false;
    }
    if n <= 3 {
        return true;
    }
    if n % 2 == 0 || n % 3 == 0 {
        return false;
    }
    let mut i = 5;
    while i * i <= n {
        if n % i == 0 || n % (i + 2) == 0 {
            return false;
        }
        i += 6;
    }
    true
}

/// Computes base^exp.
pub fn power(base: u64, exp: u32) -> u64 {
    let mut result = 1;
    let mut b = base;
    let mut e = exp;
    while e > 0 {
        if e % 2 == 1 {
            result *= b;
        }
        b *= b;
        e /= 2;
    }
    result
}
