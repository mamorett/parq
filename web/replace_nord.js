const fs = require('fs');
const path = require('path');

const srcDir = path.join(__dirname, 'src');

const replacements = {
  '--nord0': '--bg-primary',
  '--nord1': '--bg-secondary',
  '--nord2': '--border-color',
  '--nord3': '--text-muted',
  '--nord4': '--text-secondary',
  '--nord5': '--bg-tertiary',
  '--nord6': '--text-primary',
  '--nord7': '--accent-tertiary',
  '--nord8': '--accent-primary',
  '--nord9': '--accent-secondary',
  '--nord10': '--accent-dark',
  '--nord11': '--status-error',
  '--nord12': '--status-warning',
  '--nord13': '--status-attention',
  '--nord14': '--status-success',
  '--nord15': '--status-info'
};

function walk(dir) {
  fs.readdirSync(dir).forEach(file => {
    const fullPath = path.join(dir, file);
    if (fs.statSync(fullPath).isDirectory()) {
      walk(fullPath);
    } else if (fullPath.endsWith('.tsx') || fullPath.endsWith('.ts') || fullPath.endsWith('.css')) {
      let content = fs.readFileSync(fullPath, 'utf8');
      let changed = false;
      for (const [nord, semantic] of Object.entries(replacements)) {
        if (content.includes(nord)) {
          content = content.split(nord).join(semantic);
          changed = true;
        }
      }
      if (changed) {
        fs.writeFileSync(fullPath, content);
        console.log('Updated:', fullPath);
      }
    }
  });
}

walk(srcDir);
