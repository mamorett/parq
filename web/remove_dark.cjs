const fs = require('fs');
const path = require('path');
const srcDir = path.join(__dirname, 'src');

function walk(dir) {
  fs.readdirSync(dir).forEach(file => {
    const fullPath = path.join(dir, file);
    if (fs.statSync(fullPath).isDirectory()) {
      walk(fullPath);
    } else if (fullPath.endsWith('.tsx') || fullPath.endsWith('.ts')) {
      let content = fs.readFileSync(fullPath, 'utf8');
      if (content.includes('className={Classes.DARK}')) {
        content = content.split('className={Classes.DARK}').join('className="theme-editorial"');
        fs.writeFileSync(fullPath, content);
        console.log('Updated:', fullPath);
      }
    }
  });
}
walk(srcDir);
