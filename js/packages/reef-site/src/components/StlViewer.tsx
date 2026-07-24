import { useEffect, useRef, useState } from 'react';
import * as THREE from 'three';
import { STLLoader } from 'three/examples/jsm/loaders/STLLoader.js';
import { OrbitControls } from 'three/examples/jsm/controls/OrbitControls.js';

interface Props {
  url: string | null;
  pending: boolean;
  plateFits: boolean;
}

// R-4.6: three.js viewer with orbit + zoom, fed by the server preview mesh
// (R-2.6) — there is no client-side geometry generation anywhere in this
// app, only rendering of the STL the server already computed.
export default function StlViewer({ url, pending, plateFits }: Props) {
  const containerRef = useRef<HTMLDivElement>(null);
  const [loadError, setLoadError] = useState<string | null>(null);

  useEffect(() => {
    if (!url || !containerRef.current) return;
    const container = containerRef.current;
    setLoadError(null);

    const scene = new THREE.Scene();
    scene.background = new THREE.Color(0xf4ede1);

    const camera = new THREE.PerspectiveCamera(45, container.clientWidth / container.clientHeight, 0.1, 2000);
    camera.position.set(150, 150, 200);

    const renderer = new THREE.WebGLRenderer({ antialias: true });
    renderer.setSize(container.clientWidth, container.clientHeight);
    container.innerHTML = '';
    container.appendChild(renderer.domElement);

    const controls = new OrbitControls(camera, renderer.domElement);
    controls.enableDamping = true;

    scene.add(new THREE.AmbientLight(0xffffff, 0.6));
    const directional = new THREE.DirectionalLight(0xffffff, 0.8);
    directional.position.set(1, 1, 1);
    scene.add(directional);

    const loader = new STLLoader();
    let mesh: THREE.Mesh | null = null;
    let frameId: number;

    loader.load(
      url,
      (geometry) => {
        geometry.computeVertexNormals();
        geometry.center();
        const material = new THREE.MeshStandardMaterial({ color: plateFits ? 0x1f6f78 : 0xff7a59 });
        mesh = new THREE.Mesh(geometry, material);
        scene.add(mesh);

        geometry.computeBoundingBox();
        const box = geometry.boundingBox;
        if (box) {
          const size = new THREE.Vector3();
          box.getSize(size);
          const maxDim = Math.max(size.x, size.y, size.z);
          camera.position.set(maxDim * 1.2, maxDim * 1.2, maxDim * 1.5);
          controls.target.set(0, 0, 0);
        }

        const animate = () => {
          frameId = requestAnimationFrame(animate);
          controls.update();
          renderer.render(scene, camera);
        };
        animate();
      },
      undefined,
      () => setLoadError('Could not load the preview mesh.'),
    );

    const handleResize = () => {
      camera.aspect = container.clientWidth / container.clientHeight;
      camera.updateProjectionMatrix();
      renderer.setSize(container.clientWidth, container.clientHeight);
    };
    window.addEventListener('resize', handleResize);

    return () => {
      window.removeEventListener('resize', handleResize);
      cancelAnimationFrame(frameId);
      mesh?.geometry.dispose();
      renderer.dispose();
    };
  }, [url, plateFits]);

  return (
    <div className="relative rounded-lg border border-reef-teal/20 bg-reef-sand overflow-hidden" style={{ height: 360 }}>
      <div ref={containerRef} className="w-full h-full" />
      {pending && (
        <div className="absolute inset-0 flex items-center justify-center bg-reef-sand/80 text-reef-ink/70 text-sm">
          Rendering preview…
        </div>
      )}
      {!pending && !url && (
        <div className="absolute inset-0 flex items-center justify-center text-reef-ink/50 text-sm">
          Adjust the parameters to see a preview.
        </div>
      )}
      {loadError && (
        <div className="absolute inset-0 flex items-center justify-center bg-reef-sand/90 text-red-600 text-sm">
          {loadError}
        </div>
      )}
    </div>
  );
}
