mod math;
mod stats;

fn main() {
    println!("==================================================");
    println!("   Turbo Rust - Multi-File Calculation Suite Demo  ");
    println!("==================================================");

    // 1. Math Module Tests (from math.rs)
    println!("[1] Math Utilities (math.rs):");
    let n = 7;
    let fact = math::factorial(n);
    println!("    Factorial({}) = {}", n, fact);

    let (a, b) = (48, 18);
    let gcd_val = math::gcd(a, b);
    println!("    GCD({}, {})   = {}", a, b, gcd_val);

    let primes = [7, 11, 15, 19, 21, 23];
    print!("    Primes check: ");
    for &p in &primes {
        if math::is_prime(p) {
            print!("{}(Y) ", p);
        } else {
            print!("{}(N) ", p);
        }
    }
    println!();

    let (base, exp) = (2, 10);
    let p_val = math::power(base, exp);
    println!("    Power({}, {})   = {}", base, exp, p_val);

    // 2. Stats Module Tests (from stats.rs)
    println!("\n[2] Statistics Utilities (stats.rs):");
    let data = vec![12.5, 18.2, 9.4, 24.1, 15.8, 30.0, 8.6];
    let avg = stats::average(&data);
    let variance = stats::variance(&data);
    let std_dev = stats::std_dev(&data);
    let (min_val, max_val) = stats::min_max(&data);

    println!("    Dataset:  {:?}", data);
    println!("    Average:  {:.2}", avg);
    println!("    Variance: {:.2}", variance);
    println!("    StdDev:   {:.2}", std_dev);
    println!("    Min / Max: {:.2} / {:.2}", min_val, max_val);

    println!("\n[Success] All multi-file modules executed properly!");
}
