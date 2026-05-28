#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""Generate Enterprise AI Agent Architecture SVG diagram."""

lines = []

# SVG header
lines.append('<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1280 900" width="1280" height="900">')
lines.append('<style>')
lines.append("  text { font-family: 'Helvetica Neue', Helvetica, Arial, 'PingFang SC', 'Microsoft YaHei', sans-serif; }")
lines.append('  .title { font-size: 20px; font-weight: 600; fill: #111827; }')
lines.append('  .layer-title { font-size: 14px; font-weight: 600; fill: #374151; }')
lines.append('  .node-label { font-size: 13px; fill: #111827; }')
lines.append('  .sub-label { font-size: 11px; fill: #6b7280; }')
lines.append('  .legend-text { font-size: 11px; fill: #6b7280; }')
lines.append('</style>')
lines.append('<defs>')
lines.append('  <marker id="arr-blue" markerWidth="10" markerHeight="7" refX="9" refY="3.5" orient="auto">')
lines.append('    <polygon points="0 0,10 3.5,0 7" fill="#2563eb"/>')
lines.append('  </marker>')
lines.append('  <marker id="arr-green" markerWidth="10" markerHeight="7" refX="9" refY="3.5" orient="auto">')
lines.append('    <polygon points="0 0,10 3.5,0 7" fill="#16a34a"/>')
lines.append('  </marker>')
lines.append('  <marker id="arr-orange" markerWidth="10" markerHeight="7" refX="9" refY="3.5" orient="auto">')
lines.append('    <polygon points="0 0,10 3.5,0 7" fill="#ea580c"/>')
lines.append('  </marker>')
lines.append('  <marker id="arr-purple" markerWidth="10" markerHeight="7" refX="9" refY="3.5" orient="auto">')
lines.append('    <polygon points="0 0,10 3.5,0 7" fill="#9333ea"/>')
lines.append('  </marker>')
lines.append('</defs>')

# Background
lines.append('<rect width="1280" height="900" fill="#ffffff"/>')

# Title
lines.append('<text x="640" y="36" text-anchor="middle" class="title">Enterprise AI Agent Platform Architecture</text>')

# Helper function definitions as inline data
# Layer 1: Touchpoints (y=55)
ly1 = 55
lines.append(f'<rect x="30" y="{ly1}" width="1220" height="90" rx="8" fill="#eff6ff" stroke="#bfdbfe" stroke-width="1"/>')
lines.append(f'<text x="50" y="{ly1+20}" class="layer-title">Touchpoints / Channels</text>')
# Touchpoint nodes
tp_items = [("Web App", 100), ("Mobile App", 260), ("API Client", 420), ("IM/Chat", 580), ("Voice/IVR", 740), ("IoT Device", 900), ("Admin Console", 1060)]
for label, x in tp_items:
    lines.append(f'<rect x="{x}" y="{ly1+30}" width="120" height="44" rx="6" fill="#ffffff" stroke="#d1d5db" stroke-width="1.5"/>')
    lines.append(f'<text x="{x+60}" y="{ly1+56}" text-anchor="middle" class="node-label">{label}</text>')

# Layer 2: Gateway (y=160)
ly2 = 160
lines.append(f'<rect x="30" y="{ly2}" width="1220" height="90" rx="8" fill="#f0fdf4" stroke="#bbf7d0" stroke-width="1"/>')
lines.append(f'<text x="50" y="{ly2+20}" class="layer-title">API Gateway / Access Layer</text>')
gw_items = [("Auth & Token", 100), ("Rate Limit", 260), ("Protocol Adapt", 420), ("Load Balance", 580), ("Session Mgmt", 740), ("Routing Rules", 900)]
for label, x in gw_items:
    lines.append(f'<rect x="{x}" y="{ly2+30}" width="130" height="44" rx="6" fill="#ffffff" stroke="#d1d5db" stroke-width="1.5"/>')
    lines.append(f'<text x="{x+65}" y="{ly2+56}" text-anchor="middle" class="node-label">{label}</text>')

# Layer 3: Orchestration (y=265)
ly3 = 265
lines.append(f'<rect x="30" y="{ly3}" width="1220" height="100" rx="8" fill="#faf5ff" stroke="#e9d5ff" stroke-width="1"/>')
lines.append(f'<text x="50" y="{ly3+20}" class="layer-title">Orchestration / Workflow Engine</text>')
orch_items = [("Flow Engine", 100), ("Intent Router", 260), ("Context Mgr", 420), ("Multi-Agent\nCoordinator", 580), ("Fallback &\nEscalation", 740), ("Memory\nManager", 900)]
for label, x in orch_items:
    lines.append(f'<rect x="{x}" y="{ly3+30}" width="130" height="52" rx="6" fill="#ffffff" stroke="#d1d5db" stroke-width="1.5"/>')
    lbl_lines = label.split('\n')
    if len(lbl_lines) == 1:
        lines.append(f'<text x="{x+65}" y="{ly3+60}" text-anchor="middle" class="node-label">{lbl_lines[0]}</text>')
    else:
        lines.append(f'<text x="{x+65}" y="{ly3+53}" text-anchor="middle" class="node-label">{lbl_lines[0]}</text>')
        lines.append(f'<text x="{x+65}" y="{ly3+68}" text-anchor="middle" class="sub-label">{lbl_lines[1]}</text>')

# Layer 4: Agent Layer (y=380)
ly4 = 380
lines.append(f'<rect x="30" y="{ly4}" width="750" height="110" rx="8" fill="#fff7ed" stroke="#fed7aa" stroke-width="1"/>')
lines.append(f'<text x="50" y="{ly4+20}" class="layer-title">AI Agents</text>')
agent_items = [("Dialogue\nAgent", 80), ("Task\nAgent", 220), ("RAG\nAgent", 360), ("Decision\nAgent", 500), ("Code\nAgent", 640)]
for label, x in agent_items:
    lines.append(f'<rect x="{x}" y="{ly4+30}" width="120" height="60" rx="6" fill="#ffffff" stroke="#d1d5db" stroke-width="1.5"/>')
    lbl_lines = label.split('\n')
    lines.append(f'<text x="{x+60}" y="{ly4+55}" text-anchor="middle" class="node-label">{lbl_lines[0]}</text>')
    lines.append(f'<text x="{x+60}" y="{ly4+72}" text-anchor="middle" class="sub-label">{lbl_lines[1]}</text>')

# Palantir Ontology Layer (right side of agent row)
lines.append(f'<rect x="800" y="{ly4}" width="450" height="110" rx="8" fill="#fef2f2" stroke="#fecaca" stroke-width="1"/>')
lines.append(f'<text x="820" y="{ly4+20}" class="layer-title">Palantir Ontology Layer</text>')
onto_items = [("Object Types\n& Relations", 830), ("Actions &\nFunctions", 970), ("Ontology\nAPIs", 1110)]
for label, x in onto_items:
    lines.append(f'<rect x="{x}" y="{ly4+30}" width="120" height="60" rx="6" fill="#ffffff" stroke="#d1d5db" stroke-width="1.5"/>')
    lbl_lines = label.split('\n')
    lines.append(f'<text x="{x+60}" y="{ly4+55}" text-anchor="middle" class="node-label">{lbl_lines[0]}</text>')
    lines.append(f'<text x="{x+60}" y="{ly4+72}" text-anchor="middle" class="sub-label">{lbl_lines[1]}</text>')

# Layer 5: Foundation Systems (y=510)
ly5 = 510
lines.append(f'<rect x="30" y="{ly5}" width="1220" height="100" rx="8" fill="#f0fdfa" stroke="#99f6e4" stroke-width="1"/>')
lines.append(f'<text x="50" y="{ly5+20}" class="layer-title">Foundation Systems / Infrastructure</text>')
found_items = [("LLM Service\n(GPT/Claude)", 70), ("Vector DB\n(Milvus)", 230), ("Knowledge\nGraph", 390), ("Data Lake\n(S3/HDFS)", 550), ("Message Queue\n(Kafka)", 710), ("Cache\n(Redis)", 870), ("Object Store\n(MinIO)", 1030)]
for label, x in found_items:
    lines.append(f'<rect x="{x}" y="{ly5+30}" width="130" height="52" rx="6" fill="#ffffff" stroke="#d1d5db" stroke-width="1.5"/>')
    lbl_lines = label.split('\n')
    lines.append(f'<text x="{x+65}" y="{ly5+53}" text-anchor="middle" class="node-label">{lbl_lines[0]}</text>')
    lines.append(f'<text x="{x+65}" y="{ly5+68}" text-anchor="middle" class="sub-label">{lbl_lines[1]}</text>')

# Layer 6: Cross-cutting concerns (y=630) - Security, Audit, Ops
ly6 = 630

# Security
lines.append(f'<rect x="30" y="{ly6}" width="380" height="100" rx="8" fill="#fef2f2" stroke="#fecaca" stroke-width="1"/>')
lines.append(f'<text x="50" y="{ly6+20}" class="layer-title">Security</text>')
sec_items = [("IAM &\nRBAC", 60), ("Data\nEncryption", 180), ("Prompt\nFirewall", 300)]
for label, x in sec_items:
    lines.append(f'<rect x="{x}" y="{ly6+30}" width="100" height="52" rx="6" fill="#ffffff" stroke="#d1d5db" stroke-width="1.5"/>')
    lbl_lines = label.split('\n')
    lines.append(f'<text x="{x+50}" y="{ly6+53}" text-anchor="middle" class="node-label">{lbl_lines[0]}</text>')
    lines.append(f'<text x="{x+50}" y="{ly6+68}" text-anchor="middle" class="sub-label">{lbl_lines[1]}</text>')

# Audit
lines.append(f'<rect x="430" y="{ly6}" width="380" height="100" rx="8" fill="#eff6ff" stroke="#bfdbfe" stroke-width="1"/>')
lines.append(f'<text x="450" y="{ly6+20}" class="layer-title">Audit & Compliance</text>')
aud_items = [("Operation\nLog", 460), ("Decision\nTrace", 580), ("Compliance\nReport", 700)]
for label, x in aud_items:
    lines.append(f'<rect x="{x}" y="{ly6+30}" width="100" height="52" rx="6" fill="#ffffff" stroke="#d1d5db" stroke-width="1.5"/>')
    lbl_lines = label.split('\n')
    lines.append(f'<text x="{x+50}" y="{ly6+53}" text-anchor="middle" class="node-label">{lbl_lines[0]}</text>')
    lines.append(f'<text x="{x+50}" y="{ly6+68}" text-anchor="middle" class="sub-label">{lbl_lines[1]}</text>')

# Ops & Monitoring
lines.append(f'<rect x="830" y="{ly6}" width="420" height="100" rx="8" fill="#f0fdf4" stroke="#bbf7d0" stroke-width="1"/>')
lines.append(f'<text x="850" y="{ly6+20}" class="layer-title">Ops & Monitoring</text>')
ops_items = [("Metrics &\nAlerts", 860), ("Tracing\n(OpenTelemetry)", 990), ("Auto\nScaling", 1130)]
for label, x in ops_items:
    lines.append(f'<rect x="{x}" y="{ly6+30}" width="120" height="52" rx="6" fill="#ffffff" stroke="#d1d5db" stroke-width="1.5"/>')
    lbl_lines = label.split('\n')
    lines.append(f'<text x="{x+60}" y="{ly6+53}" text-anchor="middle" class="node-label">{lbl_lines[0]}</text>')
    lines.append(f'<text x="{x+60}" y="{ly6+68}" text-anchor="middle" class="sub-label">{lbl_lines[1]}</text>')

# Vertical flow arrows between layers
arrow_x_positions = [160, 420, 640, 900]
# Touchpoints -> Gateway
for ax in arrow_x_positions:
    lines.append(f'<line x1="{ax}" y1="{ly1+90}" x2="{ax}" y2="{ly2}" stroke="#2563eb" stroke-width="1.5" marker-end="url(#arr-blue)"/>')
# Gateway -> Orchestration
for ax in arrow_x_positions:
    lines.append(f'<line x1="{ax}" y1="{ly2+90}" x2="{ax}" y2="{ly3}" stroke="#2563eb" stroke-width="1.5" marker-end="url(#arr-blue)"/>')
# Orchestration -> Agents
for ax in [160, 420, 640]:
    lines.append(f'<line x1="{ax}" y1="{ly3+100}" x2="{ax}" y2="{ly4}" stroke="#2563eb" stroke-width="1.5" marker-end="url(#arr-blue)"/>')
# Orchestration -> Ontology
lines.append(f'<line x1="960" y1="{ly3+100}" x2="960" y2="{ly4}" stroke="#9333ea" stroke-width="1.5" marker-end="url(#arr-purple)"/>')
# Agents -> Foundation
for ax in [160, 420, 640]:
    lines.append(f'<line x1="{ax}" y1="{ly4+110}" x2="{ax}" y2="{ly5}" stroke="#16a34a" stroke-width="1.5" marker-end="url(#arr-green)"/>')
# Ontology -> Foundation
lines.append(f'<line x1="960" y1="{ly4+110}" x2="960" y2="{ly5}" stroke="#9333ea" stroke-width="1.5" marker-end="url(#arr-purple)"/>')

# Cross-cutting vertical dashed lines from Foundation to Security/Audit/Ops
lines.append(f'<line x1="200" y1="{ly5+100}" x2="200" y2="{ly6}" stroke="#ea580c" stroke-width="1" stroke-dasharray="4,3" marker-end="url(#arr-orange)"/>')
lines.append(f'<line x1="620" y1="{ly5+100}" x2="620" y2="{ly6}" stroke="#ea580c" stroke-width="1" stroke-dasharray="4,3" marker-end="url(#arr-orange)"/>')
lines.append(f'<line x1="1040" y1="{ly5+100}" x2="1040" y2="{ly6}" stroke="#ea580c" stroke-width="1" stroke-dasharray="4,3" marker-end="url(#arr-orange)"/>')

# Right-side vertical bar indicating cross-cutting
lines.append(f'<rect x="1260" y="{ly1}" width="6" height="680" rx="3" fill="#ea580c" opacity="0.3"/>')
lines.append(f'<text x="1258" y="{ly6+120}" text-anchor="end" class="sub-label" transform="rotate(-90,1258,{ly6+120})">Cross-Cutting Concerns</text>')

# Legend
ly_leg = 760
lines.append(f'<rect x="30" y="{ly_leg}" width="500" height="50" rx="6" fill="#f9fafb" stroke="#e5e7eb" stroke-width="1"/>')
lines.append(f'<text x="50" y="{ly_leg+18}" class="layer-title">Legend</text>')
# Blue - main flow
lines.append(f'<line x1="50" y1="{ly_leg+36}" x2="80" y2="{ly_leg+36}" stroke="#2563eb" stroke-width="1.5" marker-end="url(#arr-blue)"/>')
lines.append(f'<text x="88" y="{ly_leg+40}" class="legend-text">Request Flow</text>')
# Green - data
lines.append(f'<line x1="180" y1="{ly_leg+36}" x2="210" y2="{ly_leg+36}" stroke="#16a34a" stroke-width="1.5" marker-end="url(#arr-green)"/>')
lines.append(f'<text x="218" y="{ly_leg+40}" class="legend-text">Data Access</text>')
# Purple - ontology
lines.append(f'<line x1="310" y1="{ly_leg+36}" x2="340" y2="{ly_leg+36}" stroke="#9333ea" stroke-width="1.5" marker-end="url(#arr-purple)"/>')
lines.append(f'<text x="348" y="{ly_leg+40}" class="legend-text">Ontology Binding</text>')
# Orange - cross-cutting
lines.append(f'<line x1="460" y1="{ly_leg+36}" x2="490" y2="{ly_leg+36}" stroke="#ea580c" stroke-width="1" stroke-dasharray="4,3" marker-end="url(#arr-orange)"/>')
lines.append(f'<text x="498" y="{ly_leg+40}" class="legend-text">Cross-Cutting</text>')

# Close SVG
lines.append('</svg>')

with open('enterprise-ai-agent-architecture.svg', 'w', encoding='utf-8') as f:
    f.write('\n'.join(lines))

print("SVG generated: enterprise-ai-agent-architecture.svg")
