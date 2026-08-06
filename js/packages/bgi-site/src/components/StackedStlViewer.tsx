import { useEffect, useRef, useState } from 'react';
import * as THREE from 'three';
import { STLLoader } from 'three/examples/jsm/loaders/STLLoader.js';
import { OrbitControls } from 'three/examples/jsm/controls/OrbitControls.js';

export interface StackedTray {
  url: string;
  heightMm: number;
}

interface Props {
  trays: StackedTray[];
  pending: boolean;
  fitsBox: boolean;
}

// The one frontend piece with no reef precedent (see
// go/bgi-site/PLATFORM_FINDINGS.md): reef's StlViewer loads exactly one
// mesh. This loads N tray meshes and Y-translates each by the cumulative
// height of the trays below it, producing a real stacked-in-the-box
// preview — R-3.4's "show the customer how trays fill the box." The fit
// indicator itself (assembled height vs. box depth, with the verification
// caveat) is deliberately plain React DOM next to this component, not part
// of the 3D scene — no three.js work needed for that part.
export default function StackedStlViewer({ trays, pending, fitsBox }: Props) {
  const containerRef = useRef<HTMLDivElement>(null);
  const [loadError, setLoadError] = useState<string | null>(null);

  useEffect(() => {
    if (!containerRef.current) return;
    const container = containerRef.current;
    setLoadError(null);

    const scene = new THREE.Scene();
    scene.background = new THREE.Color(0xf7f0e3);

    const camera = new THREE.PerspectiveCamera(45, container.clientWidth / container.clientHeight, 0.1, 4000);
    camera.position.set(180, 180, 240);

    const renderer = new THREE.WebGLRenderer({ antialias: true });
    renderer.setSize(container.clientWidth, container.clientHeight);
    container.innerHTML = '';
    container.appendChild(renderer.domElement);

    const controls = new OrbitControls(camera, renderer.domElement);
    controls.enableDamping = true;

    scene.add(new THREE.AmbientLight(0xffffff, 0.65));
    const directional = new THREE.DirectionalLight(0xffffff, 0.8);
    directional.position.set(1, 1, 1);
    scene.add(directional);

    const loader = new STLLoader();
    const meshes: THREE.Mesh[] = [];
    let frameId: number;
    let cancelled = false;

    const color = fitsBox ? 0x7a5230 : 0xb5432f;

    async function loadAll() {
      let cumulativeHeightMm = 0;
      for (const tray of trays) {
        if (!tray.url) {
          cumulativeHeightMm += tray.heightMm;
          continue;
        }
        try {
          const geometry = await new Promise<THREE.BufferGeometry>((resolve, reject) => {
            loader.load(tray.url, resolve, undefined, reject);
          });
          if (cancelled) return;
          geometry.computeVertexNormals();
          geometry.center();
          const material = new THREE.MeshStandardMaterial({ color });
          const mesh = new THREE.Mesh(geometry, material);
          mesh.position.y = cumulativeHeightMm + tray.heightMm / 2;
          scene.add(mesh);
          meshes.push(mesh);
        } catch {
          setLoadError('Could not load one of the tray previews.');
        }
        cumulativeHeightMm += tray.heightMm;
      }

      if (cancelled || meshes.length === 0) return;

      const box = new THREE.Box3();
      for (const mesh of meshes) box.expandByObject(mesh);
      const size = new THREE.Vector3();
      box.getSize(size);
      const center = new THREE.Vector3();
      box.getCenter(center);
      const maxDim = Math.max(size.x, size.y, size.z, 1);
      camera.position.set(center.x + maxDim * 1.1, center.y + maxDim * 1.1, center.z + maxDim * 1.4);
      controls.target.copy(center);

      const animate = () => {
        frameId = requestAnimationFrame(animate);
        controls.update();
        renderer.render(scene, camera);
      };
      animate();
    }
    loadAll();

    const handleResize = () => {
      camera.aspect = container.clientWidth / container.clientHeight;
      camera.updateProjectionMatrix();
      renderer.setSize(container.clientWidth, container.clientHeight);
    };
    window.addEventListener('resize', handleResize);

    return () => {
      cancelled = true;
      window.removeEventListener('resize', handleResize);
      cancelAnimationFrame(frameId);
      meshes.forEach((m) => m.geometry.dispose());
      renderer.dispose();
    };
    // trays is a plain array of {url, heightMm} recreated each render from
    // API data — comparing its JSON keeps the effect from re-running (and
    // re-downloading every STL) on every unrelated parent re-render.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [JSON.stringify(trays), fitsBox]);

  const hasAnyUrl = trays.some((t) => t.url);

  return (
    <div className="card relative overflow-hidden bg-bgi-foam" style={{ height: 360 }}>
      <div ref={containerRef} className="h-full w-full" />
      {pending && (
        <div className="absolute inset-0 flex items-center justify-center bg-bgi-foam/80 text-sm text-bgi-ink/70">
          <span className="flex items-center gap-2">
            <span className="h-2 w-2 animate-shimmer rounded-full bg-bgi-teal" />
            Rendering preview…
          </span>
        </div>
      )}
      {!pending && !hasAnyUrl && (
        <div className="absolute inset-0 flex items-center justify-center text-sm text-bgi-ink/50">
          Adjust the parameters to see a preview.
        </div>
      )}
      {loadError && (
        <div className="absolute inset-0 flex items-center justify-center bg-bgi-foam/90 text-sm text-red-600">
          {loadError}
        </div>
      )}
    </div>
  );
}
