// Statistics module for Turbo Rust calculation suite

/// Computes the arithmetic mean of a slice of f64.
pub fn average(values: &[f64]) -> f64 {
    if values.is_empty() {
        return 0.0;
    }
    let total: f64 = values.iter().sum();
    total / (values.len() as f64)
}

/// Computes the sample variance of a slice of f64.
pub fn variance(values: &[f64]) -> f64 {
    if values.len() < 2 {
        return 0.0;
    }
    let avg = average(values);
    let sum_sq: f64 = values.iter().map(|v| {
        let diff = v - avg;
        diff * diff
    }).sum();
    sum_sq / ((values.len() - 1) as f64)
}

/// Computes the sample standard deviation.
pub fn std_dev(values: &[f64]) -> f64 {
    variance(values).sqrt()
}

/// Returns (min, max) from a slice of f64.
pub fn min_max(values: &[f64]) -> (f64, f64) {
    if values.is_empty() {
        return (0.0, 0.0);
    }
    let mut min = values[0];
    let mut max = values[0];
    for &v in &values[1..] {
        if v < min {
            min = v;
        }
        if v > max {
            max = v;
        }
    }
    (min, max)
}
