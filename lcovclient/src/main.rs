use std::collections::HashMap;
use std::io::{Read, Write};
use std::net::TcpStream;
use std::path::{Path, PathBuf};
use std::process::Command;
use std::time::Duration;
use anyhow::{Context, Result};
use lcov::{Reader, Record};
use std::fs;

const MAP_SIZE: usize = 65536;

#[derive(Debug)]
pub struct LcovCoverageClient {
    cov_map: [u8; MAP_SIZE],
    cov_map_total: [u8; MAP_SIZE],
    line_to_idx: HashMap<String, usize>,
    next_idx: usize,
    output_path: Option<PathBuf>,
    max_ratio: (u64, u64),
}

#[repr(u8)]
enum BlockType {
    Header = 0x01,
    CoverageInfo = 0x02,
    CmdOk = 0x03,
    SessionInfo = 0x04,
    EndOfTransmission = 0xFF,
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
        println!("[Client] Connecting to agent...");
        let mut stream = TcpStream::connect("127.0.0.1:8192")
            .context("Failed to connect to agent on port 8192")?;

        // Set read timeout to prevent hanging forever
        stream
            .set_read_timeout(Some(Duration::from_secs(5)))
            .context("Failed to set read timeout")?;

        if reset {
            println!("[Client] Sending RESET...");
            stream
                .write_all(b"RESET\n")
                .context("Failed to send RESET command")?;

            let mut ack = [0u8; 3];
            if let Err(e) = stream.read_exact(&mut ack) {
                println!("[Client] Failed to read RESET response: {e}");
            } else {
                println!("[Client] Received RESET response: {:?}", ack);
            }
            return Ok(vec![]);
        }

        println!("[Client] Requesting coverage data...");
        let mut buffer = Vec::new();

        loop {
            println!("[Client] Waiting for block...");

            let mut header = [0u8; 5];
            if let Err(e) = stream.read_exact(&mut header) {
                println!("[Client] Failed to read header: {e}");
                break;
            }

            let block_type = header[0];
            let length = u32::from_be_bytes([header[1], header[2], header[3], header[4]]) as usize;

            println!(
                "[Client] Received block type: 0x{:02X} with length {}",
                block_type, length
            );

            let mut data = vec![0u8; length];
            if let Err(e) = stream.read_exact(&mut data) {
                println!("[Client] Failed to read block data: {e}");
                break;
            }

            match block_type {
                0x01 => println!("[Client] Header block received"),
                0x02 => {
                    println!("[Client] Coverage data block received");
                    buffer.extend_from_slice(&data);
                }
                0x03 => println!("[Client] Command OK block received"),
                0x04 => println!("[Client] Session info block received"),
                0xFF => {
                    println!("[Client] End of transmission block received");
                    break;
                }
                _ => println!("[Client] Unknown block type: 0x{:02X}", block_type),
            }
        }

        println!("[Client] Finished receiving data.");
        Ok(buffer)
    }

    pub fn process_coverage(&mut self, lcov_bytes: &[u8]) -> Result<()> {
        let project_root = std::env::current_dir()?.to_string_lossy().to_string();
        let mut reader = Reader::new(lcov_bytes);
        let mut source_path = String::new();

        let mut patched_lcov = Vec::new();

        while let Some(record) = reader.next().transpose().map_err(|e| anyhow::anyhow!("Reading LCOV record failed: {}", e))? {
            match &record {
                Record::SourceFile { path } => {
                    let original = path.to_string_lossy();
                    let full_path = Path::new(&project_root).join(original.as_ref());
                    let normalized = full_path.canonicalize().unwrap_or(full_path);
                    writeln!(patched_lcov, "SF:{}", normalized.display())?;
                    source_path = normalized.display().to_string();
                }
                Record::LineData { line, count, .. } => {
                    writeln!(patched_lcov, "DA:{},{}", line, count)?;
                    if *count != 0 {
                        self.set_cov_bit(&source_path, *line, 1);
                    }
                }
                Record::EndOfRecord => {
                    writeln!(patched_lcov, "end_of_record")?;
                }
                _ => {}
            }
        }

        for (dst, src) in self.cov_map_total.iter_mut().zip(self.cov_map.iter()) {
            *dst |= *src;
        }

        fs::write("coverage.lcov", &patched_lcov)?;
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

    pub fn generate_html_report(coverage_data: &[u8]) -> Result<()> {
        fs::write("coverage.lcov", coverage_data)
            .context("Failed to write LCOV file to disk")?;

        let output = Command::new("genhtml")
            .arg("coverage.lcov")
            .arg("--output-directory")
            .arg("coverage_html")
            .output()
            .context("Failed to execute genhtml")?;

        if !output.status.success() {
            eprintln!("[Client] genhtml failed:\n{}", String::from_utf8_lossy(&output.stderr));
        } else {
            println!("[Client] HTML report generated in 'coverage_html/'");
        }

        Ok(())
    }
}


fn main() -> Result<()> {
    println!("LCOV client starting...");
    let mut client = LcovCoverageClient::new(&"127.0.0.1:8192".parse()?, None)?;
    let coverage_data = client.fetch_coverage_internal(false)?;
    if !coverage_data.is_empty() {
        client.process_coverage(&coverage_data)?;
        println!("LCOV coverage data processed successfully");
    }
    Ok(())
}

