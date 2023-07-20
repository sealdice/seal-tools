use flate2::read;
use std::error::Error;
use std::fs::File;
use std::path::{Component, Components, Path};
use std::{fs, io};
use zip::ZipArchive;

pub fn unarchive(src: &Path, target: &Path, except: Vec<&Path>) -> Result<(), Box<dyn Error>> {
    let file = File::open(src)?;
    let ext = src
        .extension()
        .ok_or("unintelligible file extension")?
        .to_str()
        .ok_or("cannot convert file extension to standard string")?
        .to_lowercase();
    let to_be_skipped = |p: &Path| except.iter().any(|e| p.starts_with(e) || e == &p);
    match ext.as_str() {
        "zip" => {
            let mut arc = ZipArchive::new(file)?;
            for i in 0..arc.len() {
                let mut file = arc.by_index(i)?;
                let name = file
                    .enclosed_name()
                    .ok_or("Warning: unsafe file containing suspicious paths")?;
                let dest = target.join(name);
                if to_be_skipped(&dest) || to_be_skipped(name) {
                    println!("skipping {:#?}", dest);
                    continue;
                }
                if let Some(parent) = dest.parent() {
                    if !parent.exists() {
                        println!("creating {:#?}", dest);
                        fs::create_dir_all(parent)?;
                    }
                }
                let mut out = File::create(&dest)?;
                println!("copying  {:#?}", dest);
                io::copy(&mut file, &mut out)?;
            }
        }
        "gz" => {
            let is_path_safe = |com: Components| {
                let normals: Vec<Component> = com
                    .into_iter()
                    .filter(|c| matches!(c, Component::Normal(_)))
                    .collect();
                !normals.is_empty()
            };
            let decoder = read::GzDecoder::new(file);
            let mut arc = tar::Archive::new(decoder);
            for entry in arc.entries()? {
                let mut file = entry?;
                let name = file.path()?;
                if !is_path_safe(name.components()) {
                    Err("Warning: unsafe file containing suspicious paths")?;
                }
                let dest = target.join(&name);
                if to_be_skipped(&dest) || to_be_skipped(&name) {
                    println!("skipping {:#?}", dest);
                    continue;
                }
                if let Some(parent) = dest.parent() {
                    if !parent.exists() {
                        println!("creating {:#?}", dest);
                        fs::create_dir_all(parent)?;
                    }
                }
                let mut out = File::create(&dest)?;
                println!("copying  {:#?}", dest);
                io::copy(&mut file, &mut out)?;
            }
        }
        _ => {
            Err(format!("unknown file extension {ext}"))?;
        }
    }
    Ok(())
}
