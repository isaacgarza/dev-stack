#!/usr/bin/env node

const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');
const yaml = require('js-yaml');

// Build the CLI first
console.log('Building dev-stack CLI...');
execSync('cd .. && go build -o docs-site/dev-stack ./cmd/dev-stack', { stdio: 'inherit' });

// Generate CLI reference
function generateCLIReference() {
    console.log('Generating CLI reference...');
    
    const helpOutput = execSync('./dev-stack --help', { encoding: 'utf8' });
    const commands = extractCommands(helpOutput);
    
    let markdown = `---
title: "CLI Reference"
description: "Complete command reference for dev-stack CLI"
lead: "Comprehensive reference for all dev-stack CLI commands and their usage"
date: ${new Date().toISOString()}
lastmod: ${new Date().toISOString()}
draft: false
weight: 50
toc: true
---

# dev-stack CLI Reference

${helpOutput.split('\n').slice(0, 3).join('\n')}

## Commands

`;

    // Generate detailed help for each command
    commands.forEach(cmd => {
        try {
            const cmdHelp = execSync(`./dev-stack ${cmd} --help`, { encoding: 'utf8' });
            markdown += `### ${cmd}\n\n\`\`\`\n${cmdHelp}\`\`\`\n\n`;
        } catch (error) {
            console.warn(`Could not get help for command: ${cmd}`);
        }
    });

    fs.writeFileSync('content/reference.md', markdown);
    console.log('✅ Generated content/reference.md');
}

// Generate services guide
function generateServicesGuide() {
    console.log('Generating services guide...');
    
    try {
        const servicesDir = '../internal/config/services';
        const services = {};
        
        // Read all service YAML files recursively
        function readServicesFromDir(dir) {
            const items = fs.readdirSync(dir);
            items.forEach(item => {
                const fullPath = path.join(dir, item);
                const stat = fs.statSync(fullPath);
                
                if (stat.isDirectory()) {
                    readServicesFromDir(fullPath);
                } else if (item.endsWith('.yaml') || item.endsWith('.yml')) {
                    const serviceName = path.basename(item, path.extname(item));
                    const content = fs.readFileSync(fullPath, 'utf8');
                    services[serviceName] = yaml.load(content);
                }
            });
        }
        
        readServicesFromDir(servicesDir);
        
        let markdown = `---
title: "Services"
description: "Available services and configuration options"
lead: "Explore all the services you can use with dev-stack"
date: ${new Date().toISOString()}
lastmod: ${new Date().toISOString()}
draft: false
weight: 30
toc: true
---

# Available Services

${Object.keys(services).length} services available for your development stack.

`;

        // Sort services by name
        const sortedServices = Object.entries(services).sort(([a], [b]) => a.localeCompare(b));
        
        sortedServices.forEach(([name, config]) => {
            markdown += `## ${name}\n\n`;
            if (config.description) {
                markdown += `${config.description}\n\n`;
            }
            if (config.defaults && config.defaults.port) {
                markdown += `**Default Port:** ${config.defaults.port}\n\n`;
            }
            if (config.docker && config.docker.services) {
                const serviceNames = Object.keys(config.docker.services);
                markdown += `**Services:** ${serviceNames.join(', ')}\n\n`;
            }
            markdown += '---\n\n';
        });

        fs.writeFileSync('content/services.md', markdown);
        console.log('✅ Generated content/services.md');
    } catch (error) {
        console.warn('Could not generate services guide:', error.message);
    }
}

// Extract command names from help output
function extractCommands(helpOutput) {
    const lines = helpOutput.split('\n');
    const commandsStart = lines.findIndex(line => line.includes('Available Commands:'));
    if (commandsStart === -1) return [];
    
    const commands = [];
    for (let i = commandsStart + 1; i < lines.length; i++) {
        const line = lines[i].trim();
        if (!line || line.startsWith('Flags:') || line.startsWith('Use ')) break;
        
        const match = line.match(/^\s*(\w+)\s+/);
        if (match && match[1] !== 'help' && match[1] !== 'completion') {
            commands.push(match[1]);
        }
    }
    return commands;
}

// Create content directory if it doesn't exist
if (!fs.existsSync('content')) {
    fs.mkdirSync('content', { recursive: true });
}

// Generate all docs
generateCLIReference();
generateServicesGuide();

// Cleanup
fs.unlinkSync('./dev-stack');
console.log('🎉 Documentation generation complete!');
