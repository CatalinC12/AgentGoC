use std::collections::HashMap;
use std::fs::{self, File};
use std::io::{Read, Write};
use std::net::TcpStream;
use std::path::{Path, PathBuf};

use anyhow::{Result, Context};
use lcov::{Reader, Record};

const MAP_SIZE: usize = 65536;

pub struct LcovCoverageClient {
    cov_map: [u8; MAP_SIZE],
    cov_map_total: [u8; MAP_SIZE],
    line_to_idx: HashMap<String, usize>,
    next_idx: usize,
    output_path: Option<PathBuf>,
    max_ratio: (u64, u64),
}

impl LcovCoverageClient {
    pub fn new(_socket_addr: &std::net::SocketAddr, output_path: Option<PathBuf>) -> Result<Self> {
        Ok(Self {
            cov_map: [0; MAP_SIZE],
            cov_map_total: [0; MAP_SIZE],
            line_to_idx: HashMap::new(),
            next_idx: 0,
            output_path,
            max_ratio: (0, 0),
        })
    }

    pub fn fetch_coverage_internal(&mut self, reset: bool) -> Result<Vec<u8>> {
        let mut stream = TcpStream::connect("127.0.0.1:8192")
            .context("Failed to connect to agent on port 8192")?;

        if reset {
            stream
                .write_all(b"RESET\n")
                .context("Failed to send RESET command")?;
            let mut ack = [0u8; 3];
            stream.read_exact(&mut ack).ok();
        }

        let mut data = Vec::new();
        stream
            .read_to_end(&mut data)
            .context("Failed to read LCOV data")?;
        Ok(data)
    }

    pub fn process_coverage(&mut self, lcov_bytes: &[u8]) -> Result<()> {
        let mut reader = Reader::new(lcov_bytes);
        let mut source_path = String::new();

        while let Some(record) = reader.next().transpose().map_err(|e| anyhow::anyhow!("Reading LCOV record failed: {}", e))? {
            match record {
                Record::SourceFile { path } => source_path = path.to_string_lossy().to_string(),
                Record::LineData { line, count, .. } if count != 0 => {
                    self.set_cov_bit(&source_path, line, 1);
                }
                _ => {}
            }
        }

        for (dst, src) in self.cov_map_total.iter_mut().zip(self.cov_map.iter()) {
            *dst |= *src;
        }

        Ok(())
    }

    fn set_cov_bit(&mut self, file: &str, line: u32, val: u8) {
        let key = format!("{}:{}", file, line);
        let idx = *self.line_to_idx.entry(key).or_insert_with(|| {
            let new_idx = self.next_idx;
            self.next_idx += 1;
            new_idx
        });
        if idx < MAP_SIZE {
            self.cov_map[idx] = val;
            self.cov_map_total[idx] = val;
        }
    }
}

fn main() -> Result<()> {
    println!("LCOV client starting...");
    let mut client = LcovCoverageClient::new(&"127.0.0.1:8192".parse()?, None)?;
    let coverage_data = client.fetch_coverage_internal(false)?;
    client.process_coverage(&coverage_data)?;
    println!("LCOV coverage data processed successfully");
    Ok(())
}
