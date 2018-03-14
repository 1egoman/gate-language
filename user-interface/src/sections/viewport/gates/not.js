export function insert(group) {
  group.append('path')
    .attr('fill', 'transparent')
    .attr('stroke', 'black')
    .attr('stroke-width', 2)

  return group
}

export function merge(group) {
  group.select('path')
    .attr('fill', d => d.active ? 'green' : 'transparent')
    .attr('d', `M17.2365571,9.47311425 L30,35 L0,35 L12.7634429,9.47311425
          C11.1248599,8.65222581 10,6.95747546 10,5 C10,2.23857625 12.2385763,0 15,0
          C17.7614237,0 20,2.23857625 20,5 C20,6.95747546 18.8751401,8.65222581
          17.2365571,9.47311425 Z`);

  return group
}
